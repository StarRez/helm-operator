package cluster

import (
	"io/ioutil"
	"os"

	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/image"
	"github.com/weaveworks/flux/policy"
	"github.com/weaveworks/flux/resource"
)

// Manifests represents how a set of files are used as definitions of
// resources, e.g., in Kubernetes, YAML files describing Kubernetes
// resources.
type Manifests interface {
	// Given a directory with manifest files, find which files define
	// which services.
	// FIXME(michael): remove when redundant
	FindDefinedServices(path string) (map[flux.ResourceID][]string, error)
	// Update the image in a manifest's bytes to that given
	UpdateImage(def []byte, resourceID flux.ResourceID, container string, newImageID image.Ref) ([]byte, error)
	// Load all the resource manifests under the path given. `baseDir`
	// is used to relativise the paths, which are supplied as absolute
	// paths to directories or files; at least one path must be
	// supplied.
	LoadManifests(baseDir, first string, rest ...string) (map[string]resource.Resource, error)
	// Parse the manifests given in an exported blob
	ParseManifests([]byte) (map[string]resource.Resource, error)
	// UpdatePolicies modifies a manifest to apply the policy update specified
	UpdatePolicies([]byte, flux.ResourceID, policy.Update) ([]byte, error)
	// ServicesWithPolicies returns all services with their associated policies
	ServicesWithPolicies(path string) (policy.ResourceMap, error)
}

// UpdateManifest looks for the manifest for a given service, reads
// its contents, applies f(contents), and writes the results back to
// the file.
func UpdateManifest(m Manifests, root string, serviceID flux.ResourceID, f func(manifest []byte) ([]byte, error)) error {
	services, err := m.FindDefinedServices(root)
	if err != nil {
		return err
	}
	paths := services[serviceID]
	if len(paths) == 0 {
		return ErrNoResourceFilesFoundForService
	}
	if len(paths) > 1 {
		return ErrMultipleResourceFilesFoundForService
	}

	def, err := ioutil.ReadFile(paths[0])
	if err != nil {
		return err
	}

	newDef, err := f(def)
	if err != nil {
		return err
	}

	fi, err := os.Stat(paths[0])
	if err != nil {
		return err
	}
	return ioutil.WriteFile(paths[0], newDef, fi.Mode())
}
