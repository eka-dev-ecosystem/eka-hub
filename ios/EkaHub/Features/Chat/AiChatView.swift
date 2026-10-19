import SwiftUI

public struct AiChatView: View {
    @State private var viewModel = AiChatViewModel()
    @State private var inputPrompt: String = ""

    public init() {}

    public var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                LazyVStack(spacing: 12) {
                    ForEach(viewModel.messages) { msg in
                        chatBubble(for: msg)
                    }
                }
                .padding(16)
            }

            // Input Bar
            HStack(spacing: 8) {
                TextField(L10n.Hub.chatPlaceholder, text: $inputPrompt)
                    .font(.subheadline)
                    .foregroundColor(.white)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 10)
                    .background(Color(red: 0.06, green: 0.09, blue: 0.16))
                    .cornerRadius(20)
                    .overlay(
                        RoundedRectangle(cornerRadius: 20)
                            .stroke(Color.gray.opacity(0.3), lineWidth: 1)
                    )

                Button(action: {
                    let text = inputPrompt
                    inputPrompt = ""
                    viewModel.sendMessage(text)
                }) {
                    Image(systemName: "paperplane.fill")
                        .foregroundColor(.white)
                        .padding(10)
                        .background(AppColors.primary)
                        .clipShape(Circle())
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 10)
            .background(AppColors.darkSurface)
        }
        .background(AppColors.darkBackground)
    }

    private func chatBubble(for msg: ChatMessageItem) -> some View {
        let isUser = msg.role == "user"
        return VStack(alignment: isUser ? .trailing : .leading, spacing: 6) {
            Text(msg.content)
                .font(.system(size: 13.5))
                .foregroundColor(.white)
                .padding(.horizontal, 14)
                .padding(.vertical, 10)
                .background(isUser ? AppColors.primary : AppColors.darkSurface)
                .cornerRadius(14)

            if !msg.actions.isEmpty {
                HStack(spacing: 8) {
                    ForEach(msg.actions) { action in
                        Button(action: { DeepLinkManager.openURL(action.deepLink) }) {
                            HStack(spacing: 4) {
                                Image(systemName: "arrow.up.right.square")
                                    .font(.caption2)
                                Text(action.title)
                                    .font(.caption)
                            }
                            .padding(.horizontal, 10)
                            .padding(.vertical, 6)
                            .background(AppColors.primary.opacity(0.2))
                            .foregroundColor(AppColors.cyan)
                            .cornerRadius(10)
                            .overlay(
                                RoundedRectangle(cornerRadius: 10)
                                    .stroke(AppColors.primary.opacity(0.4), lineWidth: 1)
                            )
                        }
                    }
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: isUser ? .trailing : .leading)
    }
}
