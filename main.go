package main

import (
	"fmt"
	"os"
	"path/filepath"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/misc"
)

func main() {
	fmt.Println("Create context")

	ctx := ocm.DefaultContext()
	cctx := ctx.CredentialsContext()

	fmt.Println("Add credentials")
	err := addCredentials(cctx, "localhost:5000", "source-user", "source-password", "./certs/1/cert.pem")
	assert(err)
	err = addCredentials(cctx, "localhost:5001", "registry-user", "registry-password", "./certs/2/cert.pem")
	assert(err)

	dir, err := os.MkdirTemp("", "")
	assert(err)

	caPath := filepath.Join(dir, "ca")

	fmt.Println("Create component version at ", caPath)
	err = buildVersion(ctx, caPath)
	assert(err)

	fmt.Println("Transfer ctf to source registry")
	c, err := comparch.Open(ctx, ctf.ACC_READONLY, caPath, 0o0777)
	assert(err)

	src, err := ctx.RepositoryForSpec(ocireg.NewRepositorySpec("localhost:5000"))
	assert(err)

	err = transfer.Transfer(c, src)
	assert(err)

	fmt.Println("Tranfer component version from source to destination registry")
	dst, err := ctx.RepositoryForSpec(ocireg.NewRepositorySpec("localhost:5001"))
	assert(err)

	cv, err := src.LookupComponentVersion("test.org/test", "1.0.0")
	assert(err)

	err = transfer.Transfer(cv, dst, standard.ResourcesByValue())
	assert(err)
}

func buildVersion(ctx ocm.Context, caPath string) error {

	ca, err := comparch.Create(ctx, accessobj.ACC_CREATE, caPath, 0o0777)
	if err != nil {
		return err
	}
	defer ca.Close()

	desc := ca.GetDescriptor()
	desc.Metadata.ConfiguredVersion = "ocm.software/v3alpha1"
	desc.Name = "test.org/test"
	desc.Version = "1.0.0"
	desc.Provider.Name = metav1.ProviderName("test.org")
	desc.CreationTime = metav1.NewTimestampP()

	if err := compdesc.Validate(desc); err != nil {
		return err
	}

	meta, err := elements.ResourceMeta("helminstaller", artifacttypes.OCI_IMAGE, elements.WithVersion("0.4.0"))
	if err != nil {
		return err
	}
	acc := ociartifact.New("ghcr.io/open-component-model/ocm/ocm.software/toi/installers/helminstaller/helminstaller:0.4.0")
	err = ca.SetResource(meta, acc, cpi.ModifyElement(true))
	if err != nil {
		return err
	}

	return nil
}

func addCredentials(ctx credentials.Context, hostname string, username string, password string, caFilePath string) error {
	id := identity.GetConsumerId(hostname, "")
	props := make(misc.Properties)
	props[identity.ATTR_USERNAME] = username
	props[identity.ATTR_PASSWORD] = password
	data, err := os.ReadFile(caFilePath)
	if err != nil {
		return err
	}
	props[identity.ATTR_CERTIFICATE_AUTHORITY] = string(data)
	cred := credentials.NewCredentials(props)
	ctx.SetCredentialsForConsumer(id, cred)
	return nil
}

func assert(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
