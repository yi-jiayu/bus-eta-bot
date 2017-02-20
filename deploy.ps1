param([string]$stage="dev")

Write-Output "Transpiling code"
npm run babel

Write-Output "Deleting dist folder"
rm -R -Force dist

Write-Output "Copying transpiled code to dist folder"
cp -R transpiled dist/transpiled

Write-Output "Copying package.json to dist folder"
node -e "var fs = require('fs'); var a = require('./package.json'); delete a.devDependencies; fs.writeFileSync('dist/package.json', JSON.stringify(a))"

Write-Output "Copying .env.json to dist folder"
cp dev.env.json dist/.env.json

Write-Output "Deploying"
gcloud alpha functions deploy "bus-eta-bot-$stage" --stage-bucket bus-eta-bot-src --entry-point main --local-path dist --memory 128MB --timeout 60 --trigger-http
