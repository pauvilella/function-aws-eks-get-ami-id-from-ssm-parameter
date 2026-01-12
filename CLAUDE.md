# CLAUDE.md

## What is this repository?

This is a **Crossplane Composition Function** written in Go. It was scaffolded from the official `function-template-go` template and is named `function-aws-eks-get-ami-id-from-ssm-parameter`.

### Current State

The codebase is still in **template/example state** - it contains the S3 bucket example from the template and needs to be replaced with the actual SSM parameter lookup logic.

**What exists now (template code):**
- `fn.go` - Contains example logic that creates S3 buckets based on an XR's `spec.names` array
- `fn_test.go` - Tests for the S3 bucket example
- `input/v1beta1/input.go` - Placeholder input struct with an `Example` field

**What needs to be built:**
- Logic to look up EKS-optimized AMI IDs from AWS SSM public parameters
- Proper input struct for configuration (SSM parameter path, region, etc.)
- Updated tests

## Project Structure

```
.
├── fn.go                 # Main function logic (RunFunction)
├── fn_test.go            # Unit tests
├── main.go               # CLI entrypoint (gRPC server)
├── input/
│   └── v1beta1/
│       └── input.go      # Function input type definition
├── package/
│   └── crossplane.yaml   # Crossplane package metadata
├── example/
│   ├── composition.yaml  # Example Composition using this function
│   ├── functions.yaml    # Function resource definition
│   └── xr.yaml           # Example XR (composite resource)
├── Dockerfile            # Multi-stage build for function container
├── go.mod / go.sum       # Go dependencies
└── .github/workflows/    # CI configuration
```

## How Crossplane Composition Functions Work

1. **Composition Functions** are called by Crossplane during the reconciliation of a Composite Resource (XR)
2. The function receives a `RunFunctionRequest` containing:
   - The observed state of the XR and any composed resources
   - The desired state from previous pipeline steps
   - Optional function-specific input configuration
3. The function returns a `RunFunctionResponse` with:
   - Updated desired state (composed resources to create/update)
   - Conditions and events to set on the XR
4. Functions run as gRPC servers, typically in containers

## Key Dependencies

- `github.com/crossplane/function-sdk-go` - SDK for writing composition functions
- `github.com/upbound/provider-aws/v2` - AWS provider types (currently S3, need SSM)
- `k8s.io/apimachinery` - Kubernetes API types

## Development Commands

```bash
# Generate code (deepcopy, CRDs)
go generate ./...

# Run tests
go test ./...

# Build Docker image
docker build . --tag=runtime

# Build Crossplane package
crossplane xpkg build -f package --embed-runtime-image=runtime
```

## Intended Functionality

The function should retrieve EKS-optimized AMI IDs from AWS SSM Parameter Store public parameters. AWS publishes these at paths like:

```
/aws/service/eks/optimized-ami/<k8s-version>/amazon-linux-2/recommended/image_id
/aws/service/eks/optimized-ami/<k8s-version>/amazon-linux-2-arm64/recommended/image_id
/aws/service/eks/optimized-ami/<k8s-version>/amazon-linux-2-gpu/recommended/image_id
```

This is useful for EKS node groups that need the latest AMI without hardcoding IDs.

## Important Notes

- **Composition functions cannot make external API calls** during execution - they are pure functions that transform state
- To get SSM parameter values, the function would need to either:
  1. Create an SSM Parameter resource and read from its status (observed state)
  2. Use a different approach like reading from XR annotations/status populated by another mechanism
- The actual AWS API call to fetch the SSM parameter value must happen through Crossplane's resource reconciliation, not directly in the function
