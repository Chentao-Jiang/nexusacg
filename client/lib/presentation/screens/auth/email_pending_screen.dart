import 'package:flutter/material.dart';
import 'package:nexusacg/presentation/screens/auth/login_screen.dart';

class EmailPendingScreen extends StatelessWidget {
  final String email;
  const EmailPendingScreen({super.key, required this.email});

  @override
  Widget build(BuildContext context) {
    return WillPopScope(
      onWillPop: () async => false,
      child: Scaffold(
        body: SafeArea(
          child: Center(
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(
                    Icons.mark_email_unread,
                    size: 80,
                    color: Color(0xFF6366F1),
                  ),
                  const SizedBox(height: 24),
                  const Text(
                    '请验证邮箱',
                    style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    '验证邮件已发送至\n$email',
                    textAlign: TextAlign.center,
                    style: const TextStyle(fontSize: 14, color: Colors.grey, height: 1.6),
                  ),
                  const SizedBox(height: 24),
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: const Color(0xFF6366F1).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Column(
                      children: [
                        Row(
                          children: [
                            Icon(Icons.info_outline, color: Color(0xFF6366F1), size: 20),
                            SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                '请打开邮箱点击"验证邮箱"按钮完成激活',
                                style: TextStyle(fontSize: 13, color: Color(0xFF6366F1)),
                              ),
                            ),
                          ],
                        ),
                        SizedBox(height: 8),
                        Row(
                          children: [
                            Icon(Icons.warning_amber, color: Color(0xFF6366F1), size: 20),
                            SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                '如未收到，请检查垃圾邮件文件夹',
                                style: TextStyle(fontSize: 13, color: Color(0xFF6366F1)),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 40),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: () {
                        Navigator.pushAndRemoveUntil(
                          context,
                          MaterialPageRoute(builder: (_) => const LoginScreen()),
                          (_) => false,
                        );
                      },
                      child: const Text('去登录'),
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
