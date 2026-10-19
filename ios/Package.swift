// swift-tools-version: 5.10
import PackageDescription

let package = Package(
    name: "EkaHub",
    defaultLocalization: "en",
    platforms: [
        .iOS(.v17),
        .macOS(.v14)
    ],
    products: [
        .library(
            name: "EkaHub",
            targets: ["EkaHub"]
        )
    ],
    dependencies: [
        .package(url: "https://github.com/google/GoogleSignIn-iOS.git", from: "7.1.0")
    ],
    targets: [
        .target(
            name: "EkaHub",
            dependencies: [
                .product(name: "GoogleSignIn", package: "GoogleSignIn-iOS"),
                .product(name: "GoogleSignInSwift", package: "GoogleSignIn-iOS")
            ],
            path: "EkaHub",
            exclude: [
                "EkaHub.entitlements",
                "Info.plist",
                "Resources/en.lproj",
                "Resources/id.lproj"
            ],
            resources: [
                .process("Resources")
            ]
        )
    ]
)
