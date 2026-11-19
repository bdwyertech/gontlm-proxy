# internal/md4

A verbatim copy of `golang.org/x/crypto/md4`, vendored from **v0.45.0**.

## Upstream source

Tag **v0.45.0**, commit **`4e0068c0098be10d7025c99ab7c50ce454c1f0f9`** (2025-11-19).

Canonical, Google-hosted:

- <https://go.googlesource.com/crypto/+/refs/tags/v0.45.0/md4/>
- <https://cs.opensource.google/go/x/crypto/+/v0.45.0:md4/md4.go>

GitHub mirror, per file:

- <https://github.com/golang/crypto/blob/v0.45.0/md4/md4.go>
- <https://github.com/golang/crypto/blob/v0.45.0/md4/md4block.go>
- <https://github.com/golang/crypto/blob/v0.45.0/md4/md4_test.go>
- <https://github.com/golang/crypto/blob/v0.45.0/md4/example_test.go>
- <https://github.com/golang/crypto/blob/v0.45.0/LICENSE> (from the module root)

## Verifying this copy

SHA-256 of the files taken unmodified from upstream:

| File | SHA-256 |
| --- | --- |
| `md4.go` | `9be73f76d4488c7a216828e4caec1d2701955f2b03b77abd019725b0442e651e` |
| `md4block.go` | `0303098d7fe87ad0ce4b8302dc984e14f1de47c53f071bbf97b808e7c81575f8` |
| `md4_test.go` | `8b13c051af844dc2d88f469a2c3b06874319cb785c25ccf64eb6134d9c2e7034` |
| `LICENSE` | `911f8f5782931320f5b8d1160a76365b83aea6447ee6c04fa6d5591467db9dad` |

Reproduce from this directory; silence means the copy is clean:

```sh
for f in md4.go md4block.go md4_test.go; do
  diff <(curl -sSfL "https://raw.githubusercontent.com/golang/crypto/v0.45.0/md4/$f") "$f"
done
diff <(curl -sSfL https://raw.githubusercontent.com/golang/crypto/v0.45.0/LICENSE) LICENSE
```

The module these files came from is attested by Go's checksum transparency log
(`https://sum.golang.org/lookup/golang.org/x/crypto@v0.45.0`, entry 46607125) as
`h1:jMBrvKuj23MTlT0bQEOBcAE0mjg8mK9RXFhRH6nyF3Q=`. That is the same hash the
`go.sum` deleted when this package was vendored had recorded, so the commit that
removed the dependency doubles as the record of what was vendored.

## Why this is vendored

NTLM computes its NT hash with MD4 (RFC 1320), so this package cannot be
swapped for a modern hash without breaking the protocol. MD4 is the *only*
thing this module ever used from `golang.org/x/crypto`.

Depending on the whole module for it was costly:

- Advisories for `golang.org/x/crypto` are recorded at **module** granularity,
  so Dependabot flagged this repository for every `x/crypto` CVE even though
  all of them were in `ssh`, `ssh/agent`, `ssh/knownhosts`, and `openpgp/*` --
  packages this module never imported. `govulncheck` correctly reported zero
  reachable vulnerabilities the entire time.
- Escaping those alerts by upgrading meant taking `x/crypto` v0.52.0, which
  declares `go 1.25` and would therefore have raised the minimum Go version
  for every consumer of this library.

Vendoring the one package we need takes this module to zero dependencies, so
neither problem can recur.

## Maintenance

Effectively none. MD4 is frozen (RFC 1320, published 1992), and the package is
deprecated upstream and will not change. `md4_test.go` and `example_test.go`
come from upstream and carry the published test vectors.

The upstream `Deprecated:` notice is kept deliberately: MD4 really is
cryptographically broken, and nothing outside this module's NTLM hashing should
use it.

## Local modifications

Only one, and it is outside the hashing code -- `example_test.go` has its import
path rewritten:

```diff
-	"golang.org/x/crypto/md4"
+	"github.com/launchdarkly/go-ntlmssp/internal/md4"
```

`md4.go`, `md4block.go`, `md4_test.go`, and `LICENSE` are byte-for-byte upstream,
so re-syncing is a plain copy plus that one-line import fix.

To keep them that way, the repository's `.golangci.yml` exempts this directory
from linting -- upstream trips `errcheck` on `hash.Write` calls, which never
return an error.

Note that `md4.go` retains upstream's `init()`, which calls
`crypto.RegisterHash(crypto.MD4, New)`. That is the only line here with an effect
outside the package, and it is kept for parity: `x/crypto/md4` registered MD4 the
same way, so vendoring changes nothing for callers. It stays safe if a consumer
also imports `x/crypto/md4`, because `RegisterHash` only assigns into an array
and both registrations install equivalent constructors.

## Licensing

Upstream is BSD-3-Clause (`LICENSE` in this directory, copied from the
`golang.org/x/crypto` module root). The copyright headers on each file are
retained. Note this differs from the MIT license covering the rest of this
repository.
