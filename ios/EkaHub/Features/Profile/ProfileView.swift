import SwiftUI

public struct ProfileView: View {
    public var onSignOut: () -> Void = {}

    public init(onSignOut: @escaping () -> Void = {}) {
        self.onSignOut = onSignOut
    }

    public var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Image(systemName: "person.crop.circle.fill")
                    .font(.system(size: 72))
                    .foregroundColor(AppColors.primary)

                VStack(spacing: 4) {
                    Text("Eka Dev")
                        .font(.title3)
                        .fontWeight(.bold)
                        .foregroundColor(.white)
                    Text("eka@eka-dev.cloud")
                        .font(.subheadline)
                        .foregroundColor(.gray)
                }

                VStack(alignment: .leading, spacing: 8) {
                    Text("Pengaturan Hub")
                        .font(.caption)
                        .foregroundColor(.gray)
                    HStack {
                        Text("Versi Aplikasi")
                            .foregroundColor(.white)
                        Spacer()
                        Text("1.0.0")
                            .foregroundColor(.gray)
                    }
                    .padding()
                    .glassCardStyle()
                }
                .padding(.horizontal)

                Spacer()

                Button(action: onSignOut) {
                    Text(L10n.Auth.signOut)
                        .foregroundColor(.red)
                        .frame(maxWidth: .infinity)
                        .padding()
                        .glassCardStyle()
                }
                .padding(.horizontal)
            }
            .padding(.top, 40)
            .background(AppColors.darkBackground)
            .navigationTitle(L10n.Nav.profile)
        }
    }
}
