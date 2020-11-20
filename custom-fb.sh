exe(){
    set -x
    "$@"
    { set +x; } 2>/dev/null
}
exe mkdir -p temp
exe cd temp
exe git clone https://github.com/phanirithvij/filebrowser
exe cd filebrowser

exe bash wizard.sh -d -b

unameOut="$(uname -s)"
case "${unameOut}" in
    MSYS*)
    exe cp filebrowser ../../"filebrowser-custom.exe";;
    CYGWIN*)
    exe cp filebrowser ../../"filebrowser-custom.exe";;
    MINGW*)
    exe cp filebrowser ../../"filebrowser-custom.exe";;
    *)
    exe cp filebrowser ../../"filebrowser-custom";;
esac

cd ../../

echo ""
echo Please add \"$PWD\" to PATH
echo ""
echo "Linux and Mac"
echo "export PATH=\$PATH:$PWD"
echo ""
echo "on Windows"
echo "set PATH=%PATH%;$PWD"
echo ""
echo "OR Install it to \$GOBIN"
case "${unameOut}" in
    MSYS*)
    echo "cp filebrowser-custom.exe $GOPATH/bin";;
    CYGWIN*)
    echo "cp filebrowser-custom.exe $GOPATH/bin";;
    MINGW*)
    echo "cp filebrowser-custom.exe $GOPATH/bin";;
    *)
    echo "cp filebrowser-custom $GOPATH/bin";;
esac
echo then run go build and ./fate
echo ""
echo "Navigate to http://localhost:3000/admin"

echo "Login: admin Password: admin"
echo "Enable user signup"
echo "Enable user home directory"
