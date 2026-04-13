# Verifying Releases

Release binaries are signed using [sigstore cosign](https://github.com/sigstore/cosign) with keyless signing.
Each signing event is recorded in the [Rekor](https://rekor.sigstore.dev) transparency log, providing a public auditable record that the binary was built by the official GitHub Actions release workflow.

## Verify a download

Install cosign:

    go install github.com/sigstore/cosign/v3/cmd/cosign@latest

Download the binary, checksums, and sigstore bundle for your platform from the [releases page](https://github.com/gokcehan/lf/releases), then run:

    cosign verify-blob lf-linux-amd64.tar.gz \
      --bundle lf-linux-amd64.tar.gz.sigstore.json \
      --certificate-identity "https://github.com/gokcehan/lf/.github/workflows/release.yml@refs/tags/TAG" \
      --certificate-oidc-issuer "https://token.actions.githubusercontent.com"

Replace `TAG` with the release tag (e.g. `r33`).

## Verify checksums

    sha256sum -c sha256sums.txt

## Reproduce a build

Builds are reproducible given the same Go version and source:

    go version -m ./lf              # shows the exact Go version used
    git checkout TAG
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.gVersion=TAG"
    sha256sum lf                    # compare with sha256sums.txt
