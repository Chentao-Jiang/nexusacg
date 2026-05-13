import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_bloc.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_state.dart';

class RegisterScreen extends StatefulWidget {
  const RegisterScreen({super.key});

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nicknameCtrl = TextEditingController();
  final _phoneCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  final _confirmCtrl = TextEditingController();

  @override
  void dispose() {
    _nicknameCtrl.dispose();
    _phoneCtrl.dispose();
    _passwordCtrl.dispose();
    _confirmCtrl.dispose();
    super.dispose();
  }

  void _handleRegister() {
    if (!_formKey.currentState!.validate()) return;
    if (_passwordCtrl.text != _confirmCtrl.text) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('两次密码输入不一致')));
      return;
    }
    context.read<AuthBloc>().add(AuthRegisterRequested(
      nickname: _nicknameCtrl.text.trim(),
      password: _passwordCtrl.text,
      phone: _phoneCtrl.text.trim().isNotEmpty ? _phoneCtrl.text.trim() : null,
    ));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('注册')),
      body: BlocListener<AuthBloc, AuthState>(
        listener: (context, state) {
          if (state is AuthUnauthenticated) {
            ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('注册成功，请登录')));
            Navigator.pop(context);
          }
          if (state is AuthError) {
            ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(state.message)));
          }
        },
        child: SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: Form(
              key: _formKey,
              child: Column(
                children: [
                  TextFormField(
                    controller: _nicknameCtrl,
                    decoration: const InputDecoration(hintText: '昵称', prefixIcon: Icon(Icons.person)),
                    validator: (v) => v == null || v.isEmpty ? '请输入昵称' : null,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _phoneCtrl,
                    decoration: const InputDecoration(hintText: '手机号（选填）', prefixIcon: Icon(Icons.phone)),
                    keyboardType: TextInputType.phone,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _passwordCtrl,
                    obscureText: true,
                    decoration: const InputDecoration(hintText: '密码', prefixIcon: Icon(Icons.lock)),
                    validator: (v) => v == null || v.length < 6 ? '密码至少6位' : null,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _confirmCtrl,
                    obscureText: true,
                    decoration: const InputDecoration(hintText: '确认密码', prefixIcon: Icon(Icons.lock_outline)),
                    validator: (v) => v == null || v != _passwordCtrl.text ? '两次密码不一致' : null,
                  ),
                  const SizedBox(height: 32),
                  BlocBuilder<AuthBloc, AuthState>(
                    builder: (context, state) {
                      return SizedBox(
                        width: double.infinity,
                        child: ElevatedButton(
                          onPressed: state is AuthLoading ? null : _handleRegister,
                          child: state is AuthLoading
                              ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                              : const Text('注册'),
                        ),
                      );
                    },
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
