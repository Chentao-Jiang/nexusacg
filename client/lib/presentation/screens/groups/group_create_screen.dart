import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:nexusacg/core/network/api_client.dart';

class GroupCreateScreen extends StatefulWidget {
  const GroupCreateScreen({super.key});
  @override
  State<GroupCreateScreen> createState() => _GroupCreateScreenState();
}

class _GroupCreateScreenState extends State<GroupCreateScreen> {
  final _nameCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  String? _coverUrl;
  bool _uploading = false;
  bool _submitting = false;

  Future<void> _pickCover() async {
    final img = await ImagePicker().pickImage(source: ImageSource.gallery, maxWidth: 800, imageQuality: 80);
    if (img == null) return;
    setState(() => _uploading = true);
    final url = await ApiClient().uploadImage(File(img.path));
    if (mounted) setState(() { _coverUrl = url; _uploading = false; });
  }

  Future<void> _create() async {
    if (_nameCtrl.text.trim().isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请输入小组名称')));
      return;
    }
    setState(() => _submitting = true);
    final res = await ApiClient().post('/groups', data: {
      'name': _nameCtrl.text.trim(),
      'description': _descCtrl.text.trim(),
      'cover_url': _coverUrl,
    });
    setState(() => _submitting = false);
    if (res.data is Map && (res.data as Map)['code'] == 0) {
      Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('创建小组')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            GestureDetector(
              onTap: _pickCover,
              child: Container(
                height: 150,
                width: double.infinity,
                decoration: BoxDecoration(
                  color: Colors.grey.shade200,
                  borderRadius: BorderRadius.circular(12),
                  image: _coverUrl != null ? DecorationImage(image: NetworkImage(_coverUrl!), fit: BoxFit.cover) : null,
                ),
                child: _uploading
                    ? const Center(child: CircularProgressIndicator())
                    : _coverUrl == null
                        ? const Center(child: Column(mainAxisSize: MainAxisSize.min, children: [Icon(Icons.add_a_photo, size: 32, color: Colors.grey), SizedBox(height: 8), Text('添加封面', style: TextStyle(color: Colors.grey))]))
                        : null,
              ),
            ),
            const SizedBox(height: 20),
            TextField(controller: _nameCtrl, decoration: const InputDecoration(labelText: '小组名称', border: OutlineInputBorder()), maxLength: 50),
            const SizedBox(height: 16),
            TextField(controller: _descCtrl, decoration: const InputDecoration(labelText: '小组简介', border: OutlineInputBorder()), maxLines: 4, maxLength: 500),
            const SizedBox(height: 24),
            SizedBox(width: double.infinity, height: 48,
              child: FilledButton(onPressed: _submitting ? null : _create, child: const Text('创建小组'))),
          ],
        ),
      ),
    );
  }
}

  @override
  void dispose() {
    _nameCtrl.dispose();
    _descCtrl.dispose();
    super.dispose();
  }
