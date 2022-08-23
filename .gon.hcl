source = [
    "./dist/psh-go-macos_darwin_all/psh-go-darwin-all",
]
bundle_id = "sh.platform.cli"

apple_id {
  username = "antonis.kalipetis@platform.sh"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Blackfire (DN59GP4LUB)"
}

zip {
  output_path = "./dist/psh-go-macos.zip"
}
