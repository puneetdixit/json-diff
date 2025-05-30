$BINARY_URL = "https://github.com/puneetdixit/json-diff/releases/latest/download/json-diff-windows-amd64.exe"
$INSTALL_PATH = "$Env:USERPROFILE\json-diff.exe"

Write-Output "Downloading json-diff for Windows..."
Invoke-WebRequest -Uri $BINARY_URL -OutFile $INSTALL_PATH
Write-Output "Downloaded json-diff to $INSTALL_PATH"
Write-Output "Add $Env:USERPROFILE to your PATH environment variable to run json-diff globally"
