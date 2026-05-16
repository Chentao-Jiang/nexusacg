import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'dart:io';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/core/repositories/repositories.dart';

class EventCreateScreen extends StatefulWidget {
  const EventCreateScreen({super.key});

  @override
  State<EventCreateScreen> createState() => _EventCreateScreenState();
}

class _EventCreateScreenState extends State<EventCreateScreen> {
  final _nameController = TextEditingController();
  final _descController = TextEditingController();
  final _addressController = TextEditingController();
  DateTime? _startTime;
  DateTime? _endTime;
  File? _coverImage;
  String? _coverUrl;
  bool _submitting = false;
  bool _uploadingImage = false;
  String? _gatedReason;

  // Whether event creation is currently available.
  // Set to false to gate until platform reaches user threshold.
  static const bool _eventCreationEnabled = false;

  @override
  void initState() {
    super.initState();
    if (!_eventCreationEnabled) {
      _gatedReason = '活动创建功能暂未开放\n平台正在积累用户，达到一定规模后将开放个人发起活动功能';
    }
  }

  Future<void> _pickCover() async {
    final picker = ImagePicker();
    final image = await picker.pickImage(source: ImageSource.gallery);
    if (image == null) return;

    setState(() => _uploadingImage = true);
    try {
      final url = await ApiClient().uploadVideo(File(image.path));
      if (url != null) {
        setState(() {
          _coverUrl = url;
          _coverImage = File(image.path);
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('封面上传失败: $e')),
        );
      }
    } finally {
      setState(() => _uploadingImage = false);
    }
  }

  Future<void> _pickDateTime(bool isStart) async {
    final date = await showDatePicker(
      context: context,
      initialDate: DateTime.now(),
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );
    if (date == null || !mounted) return;

    final time = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.now(),
    );
    if (time == null || !mounted) return;

    final dt = DateTime(date.year, date.month, date.day, time.hour, time.minute);
    setState(() {
      if (isStart) {
        _startTime = dt;
      } else {
        _endTime = dt;
      }
    });
  }

  Future<void> _submit() async {
    if (_nameController.text.trim().isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请输入活动名称')),
      );
      return;
    }
    if (_startTime == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请选择开始时间')),
      );
      return;
    }
    if (_endTime == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请选择结束时间')),
      );
      return;
    }
    if (_endTime!.isBefore(_startTime!)) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('结束时间不能早于开始时间')),
      );
      return;
    }
    if (_addressController.text.trim().isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请输入活动地址')),
      );
      return;
    }

    setState(() => _submitting = true);
    try {
      final event = await EventRepository().createEvent(
        name: _nameController.text.trim(),
        description: _descController.text.trim(),
        startTime: _startTime!,
        endTime: _endTime!,
        address: _addressController.text.trim(),
        coverUrl: _coverUrl,
      );
      if (mounted && event != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('活动创建成功')),
        );
        Navigator.pop(context);
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('创建失败，请重试')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('创建失败: $e')),
        );
      }
    } finally {
      setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_gatedReason != null) {
      return Scaffold(
        appBar: AppBar(title: const Text('创建活动')),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(32),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Icons.lock_outline,
                  size: 80,
                  color: Theme.of(context).primaryColor.withOpacity(0.3),
                ),
                const SizedBox(height: 24),
                Text(
                  _gatedReason!,
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 16, height: 1.8),
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Scaffold(
      appBar: AppBar(title: const Text('创建活动')),
      body: _submitting
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Cover image
                  GestureDetector(
                    onTap: _pickCover,
                    child: Container(
                      height: 160,
                      width: double.infinity,
                      decoration: BoxDecoration(
                        color: Colors.grey.shade100,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: Colors.grey.shade300),
                      ),
                      child: _uploadingImage
                          ? const Center(child: CircularProgressIndicator())
                          : _coverImage != null
                              ? ClipRRect(
                                  borderRadius: BorderRadius.circular(12),
                                  child: Image.file(_coverImage!, fit: BoxFit.cover),
                                )
                              : const Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Icon(Icons.add_photo_alternate, size: 40, color: Colors.grey),
                                    SizedBox(height: 8),
                                    Text('点击上传封面', style: TextStyle(color: Colors.grey)),
                                  ],
                                ),
                    ),
                  ),
                  const SizedBox(height: 20),

                  // Name
                  const Text('活动名称', style: TextStyle(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  TextField(
                    controller: _nameController,
                    decoration: const InputDecoration(
                      hintText: '例如：周六二次元随舞·人民广场',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Start time
                  const Text('开始时间', style: TextStyle(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  _DateTimePicker(
                    dateTime: _startTime,
                    onTap: () => _pickDateTime(true),
                  ),
                  const SizedBox(height: 16),

                  // End time
                  const Text('结束时间', style: TextStyle(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  _DateTimePicker(
                    dateTime: _endTime,
                    onTap: () => _pickDateTime(false),
                  ),
                  const SizedBox(height: 16),

                  // Address
                  const Text('活动地址', style: TextStyle(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  TextField(
                    controller: _addressController,
                    decoration: const InputDecoration(
                      hintText: '例如：上海市黄浦区人民广场',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Description
                  const Text('活动详情', style: TextStyle(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  TextField(
                    controller: _descController,
                    maxLines: 5,
                    decoration: const InputDecoration(
                      hintText: '描述活动内容、规则、注意事项等',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 32),

                  // Submit
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: FilledButton(
                      onPressed: _submit,
                      child: const Text('创建活动'),
                    ),
                  ),
                ],
              ),
            ),
    );
  }
}

class _DateTimePicker extends StatelessWidget {
  final DateTime? dateTime;
  final VoidCallback onTap;
  const _DateTimePicker({this.dateTime, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
        decoration: BoxDecoration(
          border: Border.all(color: Colors.grey.shade300),
          borderRadius: BorderRadius.circular(4),
        ),
        child: Row(
          children: [
            Icon(
              dateTime != null ? Icons.event : Icons.event_note,
              size: 20,
              color: dateTime != null ? null : Colors.grey,
            ),
            const SizedBox(width: 8),
            Text(
              dateTime != null ? _formatDateTime(dateTime!) : '请选择时间',
              style: TextStyle(
                color: dateTime != null ? null : Colors.grey,
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _formatDateTime(DateTime dt) {
    return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-${dt.day.toString().padLeft(2, '0')} '
        '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
  }
}
