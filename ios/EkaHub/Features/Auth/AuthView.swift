import SwiftUI

public struct AuthView: View {
    public var onAuthSuccess: () -> Void = {}

    public init(onAuthSuccess: @escaping () -> Void = {}) {
        self.onAuthSuccess = onAuthSuccess
    }

    public var body: some View {
        VStack(spacing: 20) {
            Spacer()
            Text("Personal AI Hub")
                .font(.largeTitle)
                .fontWeight(.bold)
                .foregroundColor(.white)

            Text("Satu hub untuk seluruh ekosistem aplikasi Anda")
                .font(.subheadline)
                .foregroundColor(.gray)

            Spacer()

            Button(action: onAuthSuccess) {
                Text("Masuk dengan Apple")
                    .fontWeight(.semibold)
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.white)
                    .foregroundColor(.black)
                    .cornerRadius(12)
            }
            .padding(.horizontal, 24)
            .padding(.bottom, 40)
        }
        .background(AppColors.darkBackground)
    }
}
