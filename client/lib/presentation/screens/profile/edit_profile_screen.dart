import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:image_picker/image_picker.dart';
import 'dart:io';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_bloc.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_state.dart';

class EditProfileScreen extends StatefulWidget {
  const EditProfileScreen({super.key});

  @override
  State<EditProfileScreen> createState() => _EditProfileScreenState();
}

class _EditProfileScreenState extends State<EditProfileScreen> {
  final _nicknameController = TextEditingController();
  final _bioController = TextEditingController();
  final _repo = AuthRepository();
  final _imagePicker = ImagePicker();
  String? _avatarUrl;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _loadProfile();
  }

  void _loadProfile() {
    final state = context.read<AuthBloc>().state;
    if (state is AuthAuthenticated) {
      _nicknameController.text = state.user.nickname;
      _bioController.text = state.user.bio;
      _avatarUrl = state.user.avatarUrl;
    }
  }

  bool _uploading = false;

  Future<void> _pickAvatar() async {
    final image = await _imagePicker.pickImage(source: ImageSource.gallery, maxWidth: 512, maxHeight: 512);
    if (image == null) return;

    setState(() => _uploading = true);
    try {
      final url = await ApiClient().uploadImage(File(image.path));
      if (url != null && mounted) {
        setState(() => _avatarUrl = url);
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('头像上传成功')));
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('上传失败，请重试')));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('头像上传失败: $e')));
      }
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  Future<void> _save() async {
    if (_nicknameController.text.trim().isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('昵称不能为空')));
      return;
    }

    setState(() => _saving = true);
    try {
      final result = await _repo.updateProfile(
        nickname: _nicknameController.text.trim(),
        bio: _bioController.text.trim(),
        avatarUrl: _avatarUrl,
      );
      if (mounted) {
        if (result['code'] == 0 && result['data'] != null) {
          final authBloc = context.read<AuthBloc>();
          // Refresh AuthBloc by re-fetching current user
          try {
            final token = (authBloc.state as AuthAuthenticated).accessToken;
            authBloc.add(AuthTokenRestored(token));
          } catch (_) {}
          Navigator.pop(context);
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('保存成功')));
        } else {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(result is Map && result['message'] != null ? result['message'] as String : '保存失败')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('保存失败: $e')));
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  void dispose() {
    _nicknameController.dispose();
    _bioController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('编辑资料'),
        actions: [
          TextButton(
            onPressed: _saving ? null : _save,
            child: _saving
                ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                : const Text('保存'),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            // Avatar
            GestureDetector(
              onTap: _pickAvatar,
              child: Stack(
                children: [
                  CircleAvatar(
                    radius: 50,
                    backgroundColor: _uploading ? Colors.grey.shade300 : null,
                    backgroundImage: _avatarUrl != null ? NetworkImage(_avatarUrl!) : null,
                    child: _uploading
                        ? const SizedBox(width: 24, height: 24, child: CircularProgressIndicator(strokeWidth: 2))
                        : _avatarUrl == null
                            ? const Icon(Icons.person, size: 50)
                            : null,
                  ),
                  Positioned(
                    bottom: 0,
                    right: 0,
                    child: Container(
                      padding: const EdgeInsets.all(6),
                      decoration: BoxDecoration(
                        color: Theme.of(context).colorScheme.primary,
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(Icons.camera_alt, size: 16, color: Colors.white),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),
            // Nickname
            TextField(
              controller: _nicknameController,
              decoration: const InputDecoration(
                labelText: '昵称',
                border: OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              ),
              maxLength: 20,
            ),
            const SizedBox(height: 16),
            // Bio
            TextField(
              controller: _bioController,
              decoration: const InputDecoration(
                labelText: '简介',
                border: OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                alignLabelWithHint: true,
              ),
              maxLines: 4,
              maxLength: 200,
            ),
          ],
        ),
      ),
    );
  }
}
