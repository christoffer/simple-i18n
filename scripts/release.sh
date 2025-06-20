#/bin/sh

version="$1"
if [ -z "$version" ]; then
  echo "Usage: $0 <version>"
  echo ""
  echo "Example: $0 v.0.1.0"
  exit 1
fi

if git rev-parse "$version" >/dev/null 2>&1; then
  echo "Tag $version already exists. Please use a different version."
  exit 1
fi

echo "Existing releases:"
git tag --list | grep '^v' | sort -V

read -p "Press enter release $version, or Ctrl+C to cancel"

git tag "$version"
git push origin "$version"
GOPROXY=proxy.golang.org go list -m "github.com/christoffer/simple-i18n@$version"

echo "Done"
