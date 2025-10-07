package config

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	admin "cloud.google.com/go/iam/admin/apiv1"
	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"cloud.google.com/go/storage"
	"github.com/docker/docker/client"
	"github.com/opentargets/platform-deployment-standalone/internal/tools"
	"google.golang.org/api/dns/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const gcpContextTimeout = 10 * time.Second

// ValidateNotEmpty checks if the provided string is not empty.
func ValidateNotEmpty(v string) error {
	if v == "" {
		return errors.New("cannot be empty")
	}

	return nil
}

// ValidateMaxLength checks if the provided string does not exceed the maximum length.
func ValidateMaxLength(v string, maxLength int) error {
	if len(v) > maxLength {
		return fmt.Errorf("'%s' is too long (max %d characters)", v, maxLength)
	}
	return nil
}

// ValidateRelease checks if the provided string is a valid release version.
func ValidateRelease(v string) error {
	validRelease := regexp.MustCompile(`^\d{2}.\d{2}$`)
	if !validRelease.MatchString(v) {
		return fmt.Errorf("'%s' has invalid format, it should be '25.06'", v)
	}

	// TODO: Check if DataImagePaths exist in the release URL.
	return nil
}

// ValidateURL checks if the provided string is a valid URL.
func ValidateURL(v string) error {
	if len(v) < 5 {
		return fmt.Errorf("'%s' is too short", v)
	}

	if v[:4] != "http" && v[:5] != "https" && v[:2] != "gs" {
		return fmt.Errorf("'%s' must start with http(s):// or gs://", v)
	}

	if v[len(v)-1] == '/' {
		return fmt.Errorf("'%s' must not end with a slash", v)
	}

	return ValidateMaxLength(v, 2048)
}

// ValidateImageName checks if the provided string is a valid and available docker image.
func ValidateImageName(v string) error {
	if err := ValidateNotEmpty(v); err != nil {
		return err
	}

	client, err := tools.GetDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.DistributionInspect(ctx, v, "")
	if err != nil {
		vReminder := ""
		if strings.HasPrefix(v, "v") {
			vReminder = " (remember our images are not prefixed with 'v'!)"
		}
		return fmt.Errorf("'%s' not found or not accessible: %+w %s", v, err, vReminder)
	}
	return nil
}

// ValidateVersionTag checks if the provided string is a valid and available version tag.
func ValidateVersionTag(getImageName func() string) func(v string) error {
	return func(v string) error {
		if err := ValidateNotEmpty(v); err != nil {
			return err
		}

		imageName := getImageName()
		if imageName == "" {
			return errors.New("image name is not set")
		}

		client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = client.DistributionInspect(ctx, imageName+":"+v, "")
		if err != nil {
			return fmt.Errorf("'%s' not found or not accessible", v)
		}
		return nil
	}
}

// ValidateDomainName checks if the domain name is valid.
func ValidateDomainName(v string) error {
	if err := ValidateNotEmpty(v); err != nil {
		return err
	}

	if err := ValidateMaxLength(v, 253); err != nil {
		return err
	}

	validDomain := regexp.MustCompile(`^([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$`)
	if !validDomain.MatchString(v) {
		return errors.New("must be a fully qualified domain name with lowercase letters, numbers, and hyphens")
	}

	return nil
}

// ValidateSubdomainName checks if the subdomain is valid.
func ValidateSubdomainName(v string) error {
	if err := ValidateNotEmpty(v); err != nil {
		return err
	}

	if err := ValidateMaxLength(v, 16); err != nil {
		return err
	}

	validSubdomain := regexp.MustCompile(`^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$`)
	if !validSubdomain.MatchString(v) {
		return errors.New("only single level subdomains composed of lowercase letters, numbers, hyphens, and underscores are allowed")
	}

	return nil
}

// ValidateDaysToLive checks if the number of days to live is valid.
func ValidateDaysToLive(v string) error {
	if err := ValidateNotEmpty(v); err != nil {
		return err
	}

	d, err := strconv.ParseInt(v, 10, 64)
	if err != nil || d < 0 || d > CloudDeploymentMaxDaysToLive {
		return fmt.Errorf("must be a number between 0 and %d", CloudDeploymentMaxDaysToLive)
	}

	return nil
}

// ValidateWebAppFlavor checks if the provided web app flavor is valid.
func ValidateWebAppFlavor(v string) error {
	if err := ValidateNotEmpty(v); err != nil {
		return err
	}

	validFlavors := []string{"platform", "ppp"}

	if !slices.Contains(validFlavors, v) {
		return fmt.Errorf("must be one of %s", strings.Join(validFlavors, ", "))
	}

	return nil
}

// ValidateGCPResource checks if the GCP resource name is valid.
func ValidateGCPResource(v string) error {
	if v == "" || len(v) > 63 {
		return errors.New("must be between 1 and 63 characters long")
	}

	validChars := regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
	if !validChars.MatchString(v) {
		return errors.New("can only contain lowercase letters, numbers, and hyphens, and must start with a letter")
	}
	return nil
}

// ValidateGCPProject checks if the GCP Project name exists, is valid, and accessible.
func ValidateGCPProject(v string) error {
	g := ValidateGCPResource(v)
	if g != nil {
		return g
	}

	ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
	defer cancel()

	c, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return fmt.Errorf("unable to access google cloud: %v", err)
	}

	req := &resourcemanagerpb.GetProjectRequest{
		Name: "projects/" + v,
	}

	_, err = c.GetProject(ctx, req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("'%s' does not exist", v)
		}
		if status.Code(err) == codes.InvalidArgument {
			return fmt.Errorf("'%s' is unknown", v)
		}
		if status.Code(err) == codes.PermissionDenied {
			return fmt.Errorf("'%s' is forbidden", v)
		}
		return err
	}
	return nil
}

// ValidateGCPRegion checks if the GCP Region is valid.
func ValidateGCPRegion(getGCPProject func() string) func(v string) error {
	return func(v string) error {
		g := ValidateGCPResource(v)
		if g != nil {
			return g
		}

		project := getGCPProject()
		if project == "" {
			return errors.New("gcp project is not set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
		defer cancel()

		client, err := compute.NewRegionsRESTClient(ctx)
		if err != nil {
			return fmt.Errorf("unable to access google cloud: %w", err)
		}
		defer client.Close()

		req := &computepb.GetRegionRequest{
			Project: project,
			Region:  v,
		}

		_, err = client.Get(ctx, req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("'%s' does not exist", v)
			}
			if status.Code(err) == codes.InvalidArgument {
				return fmt.Errorf("'%s' is unknown", v)
			}
			if status.Code(err) == codes.PermissionDenied {
				return fmt.Errorf("'%s' is forbidden", v)
			}
			return err
		}

		return nil
	}
}

// ValidateGCPZone checks if the GCP Zone is valid.
func ValidateGCPZone(GetGCPProject func() string) func(v string) error {
	return func(v string) error {
		g := ValidateGCPResource(v)
		if g != nil {
			return g
		}

		project := GetGCPProject()
		if project == "" {
			return errors.New("gcp project is not set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
		defer cancel()

		client, err := compute.NewZonesRESTClient(ctx)
		if err != nil {
			return fmt.Errorf("unable to access google cloud: %w", err)
		}
		defer client.Close()

		req := &computepb.GetZoneRequest{
			Project: project,
			Zone:    v,
		}

		_, err = client.Get(ctx, req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("'%s' does not exist", v)
			}
			if status.Code(err) == codes.InvalidArgument {
				return fmt.Errorf("'%s' is unknown", v)
			}
			if status.Code(err) == codes.PermissionDenied {
				return fmt.Errorf("'%s' is forbidden", v)
			}
			return err
		}
		return nil
	}
}

// ValidateGCSBucket checks if the GCS bucket exists, is valid, and accessible.
func ValidateGCSBucket(v string) error {
	if v == "" {
		return errors.New("cannot be empty")
	}

	if !strings.HasPrefix(v, "gs://") {
		return fmt.Errorf("'%s' must start with 'gs://'", v)
	}

	parts := strings.SplitN(strings.TrimPrefix(v, "gs://"), "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("'%s' must contain a bucket and an object", v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Bucket(parts[0]).Attrs(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("bucket '%s' does not exist", parts[0])
		}
		if status.Code(err) == codes.InvalidArgument {
			return fmt.Errorf("bucket '%s' is unknown", parts[0])
		}
		if status.Code(err) == codes.PermissionDenied {
			return fmt.Errorf("bucket '%s' is forbidden", parts[0])
		}
	}

	return nil
}

// ValidateGCPSnapshot checks if the GCP snapshot exists, is valid, and accessible.
func ValidateGCPSnapshot(GetGCPProject func() string) func(v string) error {
	return func(v string) error {
		g := ValidateGCPResource(v)
		if g != nil {
			return g
		}

		project := GetGCPProject()
		if project == "" {
			return errors.New("gcp project is not set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
		defer cancel()

		client, err := compute.NewSnapshotsRESTClient(ctx)
		if err != nil {
			return fmt.Errorf("unable to access google cloud: %w", err)
		}
		defer client.Close()

		req := &computepb.GetSnapshotRequest{
			Project:  project,
			Snapshot: v,
		}

		_, err = client.Get(ctx, req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("'%s' does not exist", v)
			}
			if status.Code(err) == codes.InvalidArgument {
				return fmt.Errorf("'%s' is unknown", v)
			}
			if status.Code(err) == codes.PermissionDenied {
				return fmt.Errorf("'%s' is forbidden", v)
			}
			return err
		}
		return nil
	}
}

// ValidateGCPSecret checks if the GCP secret exists, is valid, and accessible.
func ValidateGCPSecret(GetGCPProject func() string) func(v string) error {
	return func(v string) error {
		g := ValidateGCPResource(v)
		if g != nil {
			return g
		}

		project := GetGCPProject()
		if project == "" {
			return errors.New("gcp project is not set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
		defer cancel()

		client, err := secretmanager.NewRESTClient(ctx)
		if err != nil {
			return fmt.Errorf("unable to access google cloud: %w", err)
		}
		defer client.Close()

		req := &secretmanagerpb.GetSecretRequest{
			Name: fmt.Sprintf("projects/%s/secrets/%s", project, v),
		}

		_, err = client.GetSecret(ctx, req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("'%s' does not exist", v)
			}
			if status.Code(err) == codes.InvalidArgument {
				return fmt.Errorf("'%s' is unknown", v)
			}
			if status.Code(err) == codes.PermissionDenied {
				return fmt.Errorf("'%s' is forbidden", v)
			}
			return err
		}
		return nil
	}
}

// ValidateGCPCloudDNSZone checks if the GCP Cloud DNS zone exists, is valid, and accessible.
func ValidateGCPCloudDNSZone(GetGCPProject func() string) func(v string) error {
	return func(v string) error {
		g := ValidateGCPResource(v)
		if g != nil {
			return g
		}

		project := GetGCPProject()
		if project == "" {
			return errors.New("gcp project is not set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
		defer cancel()

		service, err := dns.NewService(ctx)
		if err != nil {
			return fmt.Errorf("unable to access google cloud: %v", err)
		}

		_, err = service.ManagedZones.Get(project, v).Context(ctx).Do()
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("'%s' does not exist", v)
			}
			if status.Code(err) == codes.InvalidArgument {
				return fmt.Errorf("'%s' is unknown", v)
			}
			if status.Code(err) == codes.PermissionDenied {
				return fmt.Errorf("'%s' is forbidden", v)
			}
			return err
		}
		return nil
	}
}

// ValidateGCPNetwork checks if the GCP network is safe for ppp deployments, and if it exists, is valid, and accessible.
func ValidateGCPNetwork(getWebAppFlavor func() string, GetGCPProject func() string) func(v string) error {
	return func(v string) error {
		g := ValidateGCPResource(v)
		if g != nil {
			return g
		}

		webAppFlavor := getWebAppFlavor()
		if webAppFlavor == "ppp" && v != "devinstance-ppp" {
			return errors.New("must be devinstance-ppp for ppp deployments")
		}

		project := GetGCPProject()
		if project == "" {
			return errors.New("gcp project is not set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
		defer cancel()

		client, err := compute.NewNetworksRESTClient(ctx)
		if err != nil {
			return fmt.Errorf("unable to access google cloud: %w", err)
		}
		defer client.Close()

		req := &computepb.GetNetworkRequest{
			Project: project,
			Network: v,
		}

		_, err = client.Get(ctx, req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("'%s' does not exist", v)
			}
			if status.Code(err) == codes.InvalidArgument {
				return fmt.Errorf("'%s' is unknown", v)
			}
			if status.Code(err) == codes.PermissionDenied {
				return fmt.Errorf("'%s' is forbidden", v)
			}
			return err
		}
		return nil
	}
}

// ValidateGCPServiceAccount checks if the GCP service account exists, is valid, and accessible.
func ValidateGCPServiceAccount(GetGCPProject func() string) func(v string) error {
	return func(v string) error {
		if err := ValidateNotEmpty(v); err != nil {
			return err
		}

		project := GetGCPProject()
		if project == "" {
			return errors.New("gcp project is not set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
		defer cancel()

		client, err := admin.NewIamClient(ctx)
		if err != nil {
			return fmt.Errorf("unable to access google cloud: %v", err)
		}
		defer client.Close()

		req := &adminpb.GetServiceAccountRequest{
			Name: fmt.Sprintf("projects/%s/serviceAccounts/%s", project, v),
		}

		_, err = client.GetServiceAccount(ctx, req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("'%s' does not exist", v)
			}
			if status.Code(err) == codes.InvalidArgument {
				return fmt.Errorf("'%s' is unknown", v)
			}
			if status.Code(err) == codes.PermissionDenied {
				return fmt.Errorf("'%s' is forbidden", v)
			}
			return err
		}
		return nil
	}
}
