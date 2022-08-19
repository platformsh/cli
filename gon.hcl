source = [
    "./dist/psh-go-macos_darwin_all/psh-go-darwin-all",
]
bundle_id = "sh.platform.cli"

apple_id {
  username = "antonis.kalipetis@platform.sh"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Apple Development: Antonis Kalipetis"
}
