source = [
    "./dist/platform-macos_darwin_all/platform",
]
bundle_id = "sh.platform.cli"

apple_id {
  username = "antonis.kalipetis@upsun.com"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Blackfire (DN59GP4LUB)"
}

zip {
  output_path = "./dist/platform-macos.zip"
}
