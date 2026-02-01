# Homebrew Installation

## Installing arc-to-zen via Homebrew

To install `arc-to-zen` using Homebrew:

```bash
# Add the tap
brew tap rkw6086/arc-to-zen

# Install the formula
brew install arc-to-zen
```

Or install directly in one command (auto-taps):
```bash
brew install rkw6086/arc-to-zen/arc-to-zen
```

## Updating

To update to the latest version:
```bash
brew update
brew upgrade arc-to-zen
```

## Uninstalling

```bash
brew uninstall arc-to-zen
brew untap rkw6086/arc-to-zen  # Optional: remove the tap
```

## For Maintainers

### Creating a New Release

1. Tag a new version:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. GitHub will automatically create a release tarball at:
   `https://github.com/rkw6086/arc-to-zen/archive/refs/tags/v1.0.0.tar.gz`

3. Download the tarball and calculate its SHA256:
   ```bash
   curl -L https://github.com/rkw6086/arc-to-zen/archive/refs/tags/v1.0.0.tar.gz | shasum -a 256
   ```

4. Update `Formula/arc-to-zen.rb` with the new version and SHA256:
   ```ruby
   url "https://github.com/rkw6086/arc-to-zen/archive/refs/tags/v1.0.0.tar.gz"
   sha256 "actual_sha256_hash_here"
   ```

5. Test the formula locally:
   ```bash
   brew install --build-from-source ./Formula/arc-to-zen.rb
   brew test arc-to-zen
   brew audit --strict --online arc-to-zen
   ```

6. Commit and push the updated formula.
