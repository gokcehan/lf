# Verifying Releases

Release artifacts are signed using [sigstore cosign](https://github.com/sigstore/cosign) with keyless signing.
Each signing event is recorded in the [Rekor](https://rekor.sigstore.dev) transparency log, providing a public auditable record that the artifacts were built by the official GitHub Actions release workflow.

A single signature is produced over `sha256sums.txt`, which lists the SHA-256 of every release archive. Verifying the signature on `sha256sums.txt` and then verifying each archive against `sha256sums.txt` gives the same end-to-end guarantee as a per-archive signature.

## Verify a download

Install cosign:

    go install github.com/sigstore/cosign/v3/cmd/cosign@latest

Download `sha256sums.txt`, `sha256sums.txt.sigstore.json`, and the archive(s) you want from the [releases page](https://github.com/gokcehan/lf/releases), then:

    cosign verify-blob sha256sums.txt \
      --bundle sha256sums.txt.sigstore.json \
      --certificate-identity "https://github.com/gokcehan/lf/.github/workflows/release.yml@refs/tags/TAG" \
      --certificate-oidc-issuer "https://token.actions.githubusercontent.com"

Replace `TAG` with the release tag (e.g. `r33`).

Once `sha256sums.txt` is trusted, verify the archive(s) against it:

    sha256sum --check --ignore-missing sha256sums.txt

## Reproduce a build

Builds are reproducible given the same Go version and source:

    go version -m ./lf              # shows the exact Go version used
    git checkout TAG
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.gVersion=TAG"
    sha256sum lf                    # compare with sha256sums.txt
