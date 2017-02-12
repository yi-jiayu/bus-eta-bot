param([string]$env_file='.env', [string]$out_file='out\build.zip', [switch]$v, [switch]$skip_install)

Write-Output "$(Get-Date) [START] build"

# check if out folder exists
If (!(Test-Path out)) {
Write-Output "$(Get-Date)   [START] Creating output directory (out)"

# create out folder
mkdir out | out-null

Write-Output "$(Get-Date)   [END] Creating output directory (out)"
} else {
Write-Output "$(Get-Date)   [START] Cleaning output directory (out)"

# clean out folder
if ($v) { rm -R out/* }
else { rm -R out/* | out-null }

Write-Output "$(Get-Date)   [END] Cleaning output directory (out)"
}

Write-Output "$(Get-Date)   [START] Creating build folder (out/build)"

# create out/build
if ($v) { mkdir out/build }
else { mkdir out/build | out-null }

Write-Output "$(Get-Date)   [END] Creating build folder (out/build)"

# babel into lib
Write-Output "$(Get-Date)   [START] Running babel"

if ($v) { npm run babel }
else { npm run babel | out-null }

Write-Output "$(Get-Date)   [END] Running babel"

# copy lib into out/build folder
Write-Output "$(Get-Date)   [START] Copying source files"

if ($v) { cp -R lib/* out/build }
else { cp -R lib/* out/build | out-null }

Write-Output "$(Get-Date)   [END] Copying source files"

# copy environment file into out/build folder
Write-Output "$(Get-Date)   [START] Copying environment file ($($env_file))"

if (Test-Path $env_file) {

if ($v) { cp $env_file out/build/.env }
else { cp $env_file out/build/.env | out-null }

Write-Output "$(Get-Date)   [END] Copying environment file ($($env_file))"

} else {
Write-Error "$(Get-Date)   [ERROR] Copying environment file ($($env_file)): ENOENT"
}

Write-Output "$(Get-Date)   [START] Copying package.json into out/build"

if ($v) { cp package.json out/build/package.json }
else { cp package.json out/build/package.json | out-null }

Write-Output "$(Get-Date)   [END] Copying package.json into out/build"

if ($skip_install) {
Write-Output "$(Get-Date)   [SKIPPED] Installing node modules"
} else {

Write-Output "$(Get-Date)   [START] Installing node modules"

cd out/build

Write-Output "$(Get-Date)     [START] Stepping into out/build"

# npm install into out/build folder
if ($v) { npm install --production }
else { npm install --production | out-null }

Write-Output "$(Get-Date)     [END] Stepping into out/build"

cd ../..

Write-Output "$(Get-Date)   [END] Installing node modules"

}

# zip out into out/build.zip
Write-Output "$(Get-Date)   [START] Zipping files into $out_file"

If (Test-path $out_file) {Remove-item $out_file | out-null }

Add-Type -assembly "system.io.compression.filesystem"
[io.compression.zipfile]::CreateFromDirectory('out\build', $out_file)

Write-Output "$(Get-Date)   [STOP] Zipping files into $out_file"

Write-Output "$(Get-Date) [END] build"
