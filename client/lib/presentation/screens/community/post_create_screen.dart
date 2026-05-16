import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/core/repositories/repositories.dart';

class PostCreateScreen extends StatefulWidget {
  const PostCreateScreen({super.key});

  @override
  State<PostCreateScreen> createState() => _PostCreateScreenState();
}

class _PostCreateScreenState extends State<PostCreateScreen> {
  final _titleController = TextEditingController();
  final _contentController = TextEditingController();
  final _repo = PostRepository();
  final _imagePicker = ImagePicker();
  final List<String> _imageUrls = [];
  final List<XFile> _selectedImages = [];
  bool _submitting = false;
  bool _uploadingImage = false;

  Future<void> _pickImage() async {
    final image = await _imagePicker.pickImage(source: ImageSource.gallery, maxWidth: 1920, maxHeight: 1920);
    if (image == null) return;

    setState(() {
      _selectedImages.add(image);
      _uploadingImage = true;
    });

    try {
      final url = await ApiClient().uploadImage(File(image.path));
      if (url != null && mounted) {
        setState(() => _imageUrls.add(url));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('图片上传失败')),
        );
      }
    } finally {
      if (mounted) setState(() => _uploadingImage = false);
    }
  }

  void _removeImage(int index) {
    setState(() {
      _imageUrls.removeAt(index);
      _selectedImages.removeAt(index);
    });
  }

  Future<void> _submitPost() async {
    if (_contentController.text.trim().isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请输入内容')),
      );
      return;
    }

    setState(() => _submitting = true);
    try {
      final result = await _repo.createPost(
        title: _titleController.text.trim(),
        content: _contentController.text.trim(),
        images: _imageUrls,
      );
      if (mounted) {
        if (result != null) {
          Navigator.of(context).pop();
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('发布成功，等待审核')),
          );
        } else {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('发布失败')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('发布失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  void dispose() {
    _titleController.dispose();
    _contentController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('发布帖子'),
        actions: [
          TextButton(
            onPressed: _submitting ? null : _submitPost,
            child: _submitting
                ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                : const Text('发布'),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            TextField(
              controller: _titleController,
              decoration: const InputDecoration(
                hintText: '标题（可选）',
                border: OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              ),
              maxLength: 200,
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _contentController,
              decoration: const InputDecoration(
                hintText: '分享你的想法...',
                border: OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              ),
              maxLines: 10,
              maxLength: 10000,
            ),
            const SizedBox(height: 16),
            // Image preview
            if (_imageUrls.isNotEmpty) ...[
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: List.generate(_imageUrls.length, (i) {
                  return Stack(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.network(_imageUrls[i], width: 100, height: 100, fit: BoxFit.cover),
                      ),
                      Positioned(
                        top: 2,
                        right: 2,
                        child: GestureDetector(
                          onTap: () => _removeImage(i),
                          child: Container(
                            padding: const EdgeInsets.all(2),
                            decoration: const BoxDecoration(
                              color: Colors.black54,
                              shape: BoxShape.circle,
                            ),
                            child: const Icon(Icons.close, size: 14, color: Colors.white),
                          ),
                        ),
                      ),
                    ],
                  );
                }),
              ),
              const SizedBox(height: 12),
            ],
            // Add image button
            OutlinedButton.icon(
              onPressed: _uploadingImage ? null : _pickImage,
              icon: _uploadingImage
                  ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Icon(Icons.add_photo_alternate_outlined),
              label: const Text('添加图片'),
            ),
          ],
        ),
      ),
    );
  }
}
