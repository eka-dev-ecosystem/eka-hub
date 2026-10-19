import 'package:flutter/material.dart';
import 'package:get/get.dart';

/// HubDeepLinkGuard for Flutter (e.g. F-Tracker mobile)
///
/// Handles 3 conditions:
/// a) current_uid == expected_uid: Proceed to target route
/// b) current_uid == null: Save pending URL, redirect to login with hint
/// c) current_uid != expected_uid: Show BottomSheet with 3 options (Ganti Akun, Update Hub, Batal)
class HubDeepLinkGuard extends GetMiddleware {
  static Uri? pendingDeepLink;

  @override
  RouteSettings? redirect(String? route) {
    // Current URI from app launch
    final currentUri = Uri.base;
    final expectedUid = currentUri.queryParameters['expected_uid'];

    // If no expected_uid was specified, pass through
    if (expectedUid == null || expectedUid.isEmpty) {
      return null;
    }

    // Replace with your app's actual auth storage service:
    // e.g. final storage = Get.find<StorageService>();
    final String? currentUid = _getCurrentUserUid();
    final String? currentEmail = _getCurrentUserEmail();

    // Condition B: Sedang tidak login (null)
    if (currentUid == null || currentUid.isEmpty) {
      pendingDeepLink = currentUri;
      Get.snackbar(
        'Login Diperlukan',
        'Tautan berasal dari Personal AI Hub. Silakan login untuk melanjutkan.',
        snackPosition: SnackPosition.BOTTOM,
        backgroundColor: const Color(0xFF1E293B),
        colorText: Colors.white,
      );
      return const RouteSettings(name: '/login');
    }

    // Condition A: Akun cocok persis
    if (currentUid == expectedUid) {
      return null;
    }

    // Condition C: User ganti akun di tengah jalan (current_uid != expected_uid)
    // Tampilkan BottomSheet 3 opsi
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _showAccountMismatchBottomSheet(
        expectedUid: expectedUid,
        expectedEmailHint: currentUri.queryParameters['expected_email'] ?? 'Akun Terdaftar di Hub ($expectedUid)',
        currentUid: currentUid,
        currentEmail: currentEmail ?? 'Akun Aktif Saat Ini',
        targetUri: currentUri,
      );
    });

    // Stay on current route until user decides
    return const RouteSettings(name: '/dashboard');
  }

  static void _showAccountMismatchBottomSheet({
    required String expectedUid,
    required String expectedEmailHint,
    required String currentUid,
    required String currentEmail,
    required Uri targetUri,
  }) {
    Get.bottomSheet(
      Container(
        padding: const EdgeInsets.all(24),
        decoration: const BoxDecoration(
          color: Color(0xFF1E293B),
          borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Row(
              children: [
                Icon(Icons.warning_amber_rounded, color: Color(0xFFF59E0B), size: 28),
                SizedBox(width: 10),
                Text(
                  'Akun Tidak Sesuai',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            const Text(
              'Tautan dari Hub meminta akun:',
              style: TextStyle(color: Color(0xFF94A3B8), fontSize: 13),
            ),
            Text(
              expectedEmailHint,
              style: const TextStyle(
                color: Color(0xFF38BDF8),
                fontWeight: FontWeight.w600,
                fontSize: 14,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              'Sedangkan akun yang sedang aktif saat ini:',
              style: TextStyle(color: Color(0xFF94A3B8), fontSize: 13),
            ),
            Text(
              currentEmail,
              style: const TextStyle(
                color: Color(0xFF10B981),
                fontWeight: FontWeight.w600,
                fontSize: 14,
              ),
            ),
            const SizedBox(height: 24),

            // Opsi 1: Ganti ke Akun yang diminta
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF4F46E5),
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
                onPressed: () {
                  Get.back(); // close sheet
                  pendingDeepLink = targetUri;
                  // Logout user and route to login
                  _logoutUser();
                  Get.offAllNamed('/login');
                },
                child: const Text('1. Ganti ke Akun Terkait Hub', style: TextStyle(color: Colors.white, fontWeight: FontWeight.w600)),
              ),
            ),
            const SizedBox(height: 8),

            // Opsi 2: Perbarui Koneksi Hub ke Akun Saat Ini
            SizedBox(
              width: double.infinity,
              child: OutlinedButton(
                style: OutlinedButton.styleFrom(
                  foregroundColor: const Color(0xFF38BDF8),
                  side: const BorderSide(color: Color(0xFF38BDF8)),
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
                onPressed: () async {
                  Get.back();
                  // Call Hub API PUT /api/v1/hub/link/update
                  await _updateHubConnection(appId: 'ftracker', currentUid: currentUid, currentEmail: currentEmail);
                  // Proceed to target route
                  Get.toNamed(targetUri.path, parameters: targetUri.queryParameters);
                },
                child: const Text('2. Perbarui Koneksi Hub ke Akun Ini', style: TextStyle(fontWeight: FontWeight.w600)),
              ),
            ),
            const SizedBox(height: 8),

            // Opsi 3: Batal
            SizedBox(
              width: double.infinity,
              child: TextButton(
                onPressed: () => Get.back(),
                child: const Text('3. Batal (Tetap di Akun Saat Ini)', style: TextStyle(color: Color(0xFF94A3B8))),
              ),
            ),
          ],
        ),
      ),
      isDismissible: true,
      enableDrag: true,
    );
  }

  static String? _getCurrentUserUid() {
    // Return logged in user's UID or null if unauthenticated
    return 'usr_ftracker_active';
  }

  static String? _getCurrentUserEmail() {
    return 'active@email.com';
  }

  static void _logoutUser() {
    // Clear tokens in storage
  }

  static Future<void> _updateHubConnection({
    required String appId,
    required String currentUid,
    required String currentEmail,
  }) async {
    // PUT /api/v1/hub/link/update
  }
}
