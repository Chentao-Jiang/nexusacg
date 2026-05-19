import 'dart:async';
import 'package:flutter/material.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/screens/auth/login_screen.dart';

class EmailPendingScreen extends StatefulWidget {
  final String email;
  const EmailPendingScreen({super.key, required this.email});

  @override
  State<EmailPendingScreen> createState() => _EmailPendingScreenState();
}

class _EmailPendingScreenState extends State<EmailPendingScreen> {
  final _authRepo = AuthRepository();
  Timer? _pollTimer;
  bool _checking = false;
  bool _resending = false;

  @override
  void initState() {
    super.initState();
    _pollTimer = Timer.periodic(const Duration(seconds: 10), (_) => _checkStatus());
    Future.delayed(const Duration(seconds: 5), _checkStatus);
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    super.dispose();
  }

  Future<void> _checkStatus() async {
    if (_checking) return;
    setState(() => _checking = true);
    try {
      final result = await _authRepo.getCurrentUser();
      if (result['code'] == 0 && result['data'] != null) {
        _pollTimer?.cancel();
        if (mounted) _goToLogin();
      }
    } catch (_) {
      // Still pending
    } finally {
      if (mounted) setState(() => _checking = false);
    }
  }

  void _goToLogin() {
    _pollTimer?.cancel();
    Navigator.pushAndRemoveUntil(
      context,
      MaterialPageRoute(builder: (_) => const LoginScreen()),
      (_) => false,
    );
  }

  Future<void> _resendEmail() async {
    setState(() => _resending = true);
    try {
      await _authRepo.resendEmailVerification(widget.email);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('验证邮件已重新发送，请查收')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('发送失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _resending = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      child: Scaffold(
        body: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(32),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.mark_email_unread, size: 80, color: Color(0xFF6366F1)),
                  const SizedBox(height: 24),
                  const Text('请验证邮箱', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 12),
                  Text('验证邮件已发送至\n${widget.email}',
                      textAlign: TextAlign.center,
                      style: const TextStyle(fontSize: 14, color: Colors.grey, height: 1.6)),
                  const SizedBox(height: 24),
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: const Color(0xFF6366F1).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Column(
                      children: [
                        Row(children: [
                          Icon(Icons.info_outline, color: Color(0xFF6366F1), size: 20),
                          SizedBox(width: 8),
                          Expanded(child: Text('请打开邮箱点击"验证邮箱"按钮完成激活',
                              style: TextStyle(fontSize: 13, color: Color(0xFF6366F1)))),
                        ]),
                        SizedBox(height: 8),
                        Row(children: [
                          Icon(Icons.warning_amber, color: Color(0xFF6366F1), size: 20),
                          SizedBox(width: 8),
                          Expanded(child: Text('如未收到，请检查垃圾邮件文件夹',
                              style: TextStyle(fontSize: 13, color: Color(0xFF6366F1)))),
                        ]),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                  if (_checking)
                    const Padding(
                      padding: EdgeInsets.symmetric(vertical: 8),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(width: 14, height: 14, child: CircularProgressIndicator(strokeWidth: 2)),
                          SizedBox(width: 8),
                          Text('正在检查验证状态...', style: TextStyle(fontSize: 13, color: Colors.grey)),
                        ],
                      ),
                    ),
                  const SizedBox(height: 16),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton.icon(
                      onPressed: _goToLogin,
                      icon: const Icon(Icons.login),
                      label: const Text('去登录'),
                    ),
                  ),
                  const SizedBox(height: 12),
                  TextButton(
                    onPressed: _checking ? null : _checkStatus,
                    child: const Text('检查验证状态'),
                  ),
                  const SizedBox(height: 8),
                  SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      onPressed: _resending ? null : _resendEmail,
                      icon: _resending
                          ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                          : const Icon(Icons.refresh),
                      label: const Text('重新发送验证邮件'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
