import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_bloc.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_state.dart';
import 'package:nexusacg/presentation/screens/auth/email_register_screen.dart';
import 'package:nexusacg/presentation/screens/auth/register_screen.dart';
import 'package:nexusacg/presentation/screens/auth/qq_oauth_screen.dart';
import 'package:nexusacg/core/repositories/repositories.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _phoneCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  bool _obscure = true;

  @override
  void dispose() {
    _phoneCtrl.dispose();
    _passwordCtrl.dispose();
    super.dispose();
  }

  void _handleLogin() {
    if (!_formKey.currentState!.validate()) return;
    context.read<AuthBloc>().add(AuthLoginRequested(
      phone: _phoneCtrl.text.trim(),
      password: _passwordCtrl.text,
    ));
  }

  Future<void> _handleQQLogin() async {
    try {
      final authUrl = await AuthRepository().getQQAuthUrl();
      if (authUrl != null && mounted) {
        await Navigator.push(
          context,
          MaterialPageRoute(builder: (_) => QQOAuthLoginScreen(authUrl: authUrl)),
        );
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('QQ 登录未配置，请联系管理员')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('获取 QQ 授权链接失败: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: BlocListener<AuthBloc, AuthState>(
        listener: (context, state) {
          if (state is AuthError) {
            ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(state.message)));
          }
        },
        child: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Text('次元链', style: TextStyle(fontSize: 36, fontWeight: FontWeight.bold, color: Color(0xFF6366F1))),
                  const SizedBox(height: 8),
                  const Text('NexusACG', style: TextStyle(fontSize: 16, color: Colors.grey)),
                  const SizedBox(height: 40),
                  Form(
                    key: _formKey,
                    child: Column(
                      children: [
                        TextFormField(
                          controller: _phoneCtrl,
                          decoration: const InputDecoration(
                            hintText: '手机号',
                            prefixIcon: Icon(Icons.phone),
                          ),
                          keyboardType: TextInputType.phone,
                          validator: (v) => v == null || v.isEmpty ? '请输入手机号' : null,
                        ),
                        const SizedBox(height: 16),
                        TextFormField(
                          controller: _passwordCtrl,
                          obscureText: _obscure,
                          decoration: InputDecoration(
                            hintText: '密码',
                            prefixIcon: const Icon(Icons.lock),
                            suffixIcon: IconButton(
                              icon: Icon(_obscure ? Icons.visibility_off : Icons.visibility),
                              onPressed: () => setState(() => _obscure = !_obscure),
                            ),
                          ),
                          validator: (v) => v == null || v.length < 6 ? '密码至少6位' : null,
                        ),
                        const SizedBox(height: 24),
                        BlocBuilder<AuthBloc, AuthState>(
                          builder: (context, state) {
                            return SizedBox(
                              width: double.infinity,
                              child: ElevatedButton(
                                onPressed: state is AuthLoading ? null : _handleLogin,
                                child: state is AuthLoading
                                    ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                                    : const Text('登录'),
                              ),
                            );
                          },
                        ),
                        const SizedBox(height: 16),
                        TextButton(
                          onPressed: () {
                            showModalBottomSheet(
                              context: context,
                              builder: (_) => Column(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  ListTile(
                                    leading: const Icon(Icons.email),
                                    title: const Text('邮箱注册'),
                                    onTap: () {
                                      Navigator.pop(context);
                                      Navigator.push(context, MaterialPageRoute(builder: (_) => const EmailRegisterScreen()));
                                    },
                                  ),
                                  ListTile(
                                    leading: const Icon(Icons.phone),
                                    title: const Text('手机注册'),
                                    onTap: () {
                                      Navigator.pop(context);
                                      Navigator.push(context, MaterialPageRoute(builder: (_) => const RegisterScreen()));
                                    },
                                  ),
                                ],
                              ),
                            );
                          },
                          child: const Text('还没有账号？立即注册'),
                        ),
                        const SizedBox(height: 16),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            _socialButton(Icons.chat, '微信登录', () {}),
                            const SizedBox(width: 16),
                            _socialButton(Icons.person, 'QQ登录', _handleQQLogin),
                          ],
                        ),
                      ],
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

  Widget _socialButton(IconData icon, String label, VoidCallback onTap) {
    return OutlinedButton.icon(
      onPressed: onTap,
      icon: Icon(icon, size: 20),
      label: Text(label),
      style: OutlinedButton.styleFrom(padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12)),
    );
  }
}
