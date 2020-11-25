exe(){
    set -x
    "$@"
    { set +x; } 2>/dev/null
}

ASSETS="false"
RICE="false"

debugInfo () {
  echo "Repo:           $REPO"
  echo "Build assets:   $ASSETS"
  echo "Rice assets:    $RICE"
}

buildAssets () {
  exe cd $REPO
  exe rm -rf frontend/dist

  exe cd $REPO/frontend

  if [ "$CI" = "true" ]; then
    exe npm ci
  else
    exe npm install
  fi

  exe npm run build
}

riceAssets () {
  if ! [ -x "$(command -v rice)" ]; then
    exe go install github.com/GeertJohan/go.rice/rice
  fi
  exe rm -f http/rice-box.go

  exe cd $REPO/http
  exe rm -rf rice-box.go
  exe rice embed-go -i "github.com/filebrowser/filebrowser/v2/http"
}


REPO=$(readlink -f "filebrowser")

usage() {
  echo "Usage: $0 [-a assets only] [-r rice only] [-d debug]" 1>&2;
  exit 1;
}

DEBUG="false"

while getopts "ar:d" o; do
  case "${o}" in
    a)
      ASSETS="true"
      ;;
    r)
      RICE="true"
      ;;
    d)
      DEBUG="true"
      ;;
    *)
      usage
      ;;
  esac
done
shift $((OPTIND-1))

if [ "$DEBUG" = "true" ]; then
  debugInfo
fi

if [ "$ASSETS" = "true" ]; then
  buildAssets
fi

if [ "$RICE" = "true" ]; then
  riceAssets
fi
