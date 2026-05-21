import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/core/repositories/repositories.dart';

/// Tracks an image being uploaded: local file + remote URL
class _UploadedImage {
  final File file;
  String? url;
  bool uploading;

  _UploadedImage({required this.file, this.url, this.uploading = false});
}

class PostCreateScreen extends StatefulWidget {
  final String? groupId;
  const PostCreateScreen({super.key, this.groupId});

  @override
  State<PostCreateScreen> createState() => _PostCreateScreenState();
}

class _PostCreateScreenState extends State<PostCreateScreen> {
  final _titleController = TextEditingController();
  final _contentController = TextEditingController();
  final _repo = PostRepository();
  final _imagePicker = ImagePicker();
  final List<_UploadedImage> _images = [];
  String? _videoUrl;
  String _visibility = 'public'; // public | followers | private
  bool _submitting = false;
  bool _uploadingAny = false;
  bool _uploadingVideo = false;
  double _videoUploadProgress = 0.0;

  Future<void> _pickImage() async {
    final image = await _imagePicker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 1920,
      maxHeight: 1920,
      imageQuality: 85,
    );
    if (image == null) return;

    final item = _UploadedImage(file: File(image.path), uploading: true);
    setState(() {
      _images.add(item);
      _uploadingAny = true;
    });

    try {
      final url = await ApiClient().uploadImage(item.file);
      if (mounted) {
        setState(() {
          item.url = url;
          item.uploading = false;
          _uploadingAny = _images.any((i) => i.uploading);
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          item.uploading = false;
          _uploadingAny = _images.any((i) => i.uploading);
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('图片上传失败: $e')),
        );
      }
    }
  }

  void _removeImage(int index) {
    setState(() => _images.removeAt(index));
  }

  Future<void> _pickVideo() async {
    final video = await _imagePicker.pickVideo(
      source: ImageSource.gallery,
      maxDuration: const Duration(minutes: 5),
    );
    if (video == null) return;

    final file = File(video.path);

    // Check file size before upload (server limit: 200MB)
    final fileSize = await file.length();
    if (fileSize > 200 * 1024 * 1024) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('视频过大 (' + (fileSize ~/ (1024 * 1024)).toString() + 'MB)，上限 200MB')),
        );
      }
      return;
    }

    await _doUpload(file);
  }

  Future<void> _doUpload(File videoFile) async {
    setState(() {
      _uploadingVideo = true;
      _uploadingAny = true;
      _videoUploadProgress = 0.0;
    });

    try {
      final result = await ApiClient().uploadVideo(
        videoFile,
        onProgress: (sent, total) {
          if (mounted && total > 0) {
            setState(() => _videoUploadProgress = sent / total);
          }
        },
      );

      if (!mounted) return;

      if (result.isSuccess) {
        setState(() {
          _videoUrl = result.url;
        });

        setState(() {
          _uploadingVideo = false;
          _uploadingAny = false;
        });
      } else {
        setState(() {
          _uploadingVideo = false;
          _uploadingAny = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('视频上传失败：${result.error}')),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _uploadingVideo = false;
          _uploadingAny = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('视频上传失败: $e')),
        );
      }
    }
  }


  void _removeVideo() {
    setState(() {
      _videoUrl = null;
      _uploadingVideo = false;
      _videoUploadProgress = 0.0;
    });
  }

  Future<void> _submitPost() async {
    if (_contentController.text.trim().isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请输入内容')),
      );
      return;
    }

    // Wait for uploads
    if (_uploadingVideo || _images.any((i) => i.uploading)) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请等待上传完成')),
      );
      return;
    }

    setState(() => _submitting = true);
    try {
      final imageUrls = _images.map((i) => i.url).whereType<String>().toList();
final result = await _repo.createPost(
        title: _titleController.text.trim(),
        content: _contentController.text.trim(),
        images: imageUrls,
        videoUrl: _videoUrl,
        visibility: _visibility,
        groupId: widget.groupId,
      );
      if (mounted) {
        if (result != null) {
          Navigator.of(context).pop();
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('发布成功')),
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
            // Visibility selector
            Row(
              children: [
                const Icon(Icons.visibility_outlined, size: 20),
                const SizedBox(width: 8),
                const Text('可见范围', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w500)),
                const Spacer(),
                DropdownButton<String>(
                  value: _visibility,
                  underline: const SizedBox(),
                  items: const [
                    DropdownMenuItem(value: 'public', child: Text('所有人')),
                    DropdownMenuItem(value: 'followers', child: Text('粉丝可见')),
                    DropdownMenuItem(value: 'private', child: Text('仅自己')),
                  ],
                  onChanged: (v) {
                    if (v != null) setState(() => _visibility = v);
                  },
                ),
              ],
            ),
            const SizedBox(height: 16),
            // Video upload progress / preview
            if (_uploadingVideo) ...[
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.blue.shade50,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.blue.shade200),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const SizedBox(
                          width: 20, height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2.5),
                        ),
                        const SizedBox(width: 10),
                        Text(
                          '上传中 ${(_videoUploadProgress * 100).toStringAsFixed(0)}%',
                          style: const TextStyle(color: Colors.blue, fontWeight: FontWeight.w500),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(4),
                      child: LinearProgressIndicator(
                        value: _videoUploadProgress,
                        minHeight: 6,
                        backgroundColor: Colors.blue.shade100,
                        valueColor: const AlwaysStoppedAnimation<Color>(Colors.blue),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 12),
            ],
            // Video uploaded (not uploading)
            if (_videoUrl != null && !_uploadingVideo) ...[
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.blue.shade50,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.blue.shade200),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.video_file, color: Colors.blue),
                    const SizedBox(width: 8),
                    const Expanded(child: Text('视频已上传', style: TextStyle(color: Colors.blue))),
                    IconButton(
                      icon: const Icon(Icons.close, size: 20),
                      onPressed: _removeVideo,
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 12),
            ],
            // Image previews
            if (_images.isNotEmpty) ...[
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: List.generate(_images.length, (i) {
                  final item = _images[i];
                  return Stack(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.file(
                          item.file,
                          width: 100,
                          height: 100,
                          fit: BoxFit.cover,
                          errorBuilder: (_, __, ___) => Container(
                            width: 100, height: 100, color: Colors.grey.shade200,
                            child: const Icon(Icons.broken_image),
                          ),
                        ),
                      ),
                      if (item.uploading)
                        Container(
                          width: 100, height: 100,
                          color: Colors.black54,
                          child: const Center(
                            child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
                          ),
                        ),
                      Positioned(
                        top: 2, right: 2,
                        child: GestureDetector(
                          onTap: () => _removeImage(i),
                          child: Container(
                            padding: const EdgeInsets.all(2),
                            decoration: const BoxDecoration(
                              color: Colors.black54, shape: BoxShape.circle,
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
            // Action buttons
            Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _uploadingAny ? null : _pickImage,
                    icon: _images.any((i) => i.uploading)
                        ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Icon(Icons.add_photo_alternate_outlined),
                    label: const Text('添加图片'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _uploadingAny || _videoUrl != null ? null : _pickVideo,
                    icon: _uploadingAny
                        ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Icon(Icons.videocam_outlined),
                    label: const Text('添加视频'),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
