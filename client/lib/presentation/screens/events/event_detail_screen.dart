import 'package:flutter/material.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/network/api_client.dart';

class EventDetailScreen extends StatelessWidget {
  final EventModel event;
  const EventDetailScreen({super.key, required this.event});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isPast = event.endTime.isBefore(DateTime.now());
    final statusText = _getStatusText();

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 200,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              title: Text(event.name, style: const TextStyle(fontSize: 16)),
              background: event.coverUrl != null
                  ? Image.network(event.coverUrl!, fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => _buildDefaultCover(theme))
                  : _buildDefaultCover(theme),
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.share),
                onPressed: () => _showShareSheet(context),
              ),
            ],
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(children: [
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                      decoration: BoxDecoration(
                        color: isPast ? Colors.grey.shade200 : theme.primaryColor.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(statusText, style: TextStyle(
                        color: isPast ? Colors.grey : theme.primaryColor, fontWeight: FontWeight.w500)),
                    ),
                  ]),
                  const SizedBox(height: 16),
                  Text(event.name, style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 20),
                  _InfoCard(icon: Icons.calendar_today, label: '时间',
                    value: '${_formatDate(event.startTime)} ${_formatTime(event.startTime)} - ${_formatTime(event.endTime)}'),
                  const SizedBox(height: 8),
                  _InfoCard(icon: Icons.location_on, label: '地点', value: event.address),
                  const SizedBox(height: 24), const Divider(), const SizedBox(height: 16),
                  const Text('活动详情', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  Text(event.description.isEmpty ? '暂无详情' : event.description,
                    style: TextStyle(fontSize: 15, color: event.description.isEmpty ? Colors.grey : null, height: 1.6)),
                  const SizedBox(height: 32),
                  if (!isPast) _RegisterButton(event: event),
                  const SizedBox(height: 32),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDefaultCover(ThemeData theme) {
    return Container(
      color: theme.primaryColor.withOpacity(0.1),
      child: Center(child: Icon(Icons.event, size: 64, color: theme.primaryColor.withOpacity(0.5))),
    );
  }

  static void _showShareSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      builder: (ctx) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 20),
          child: Wrap(children: [
            ListTile(
              leading: const Icon(Icons.copy, color: Colors.blue),
              title: const Text('复制链接'),
              onTap: () {
                Navigator.pop(ctx);
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('链接已复制')));
              },
            ),
            ListTile(
              leading: const Icon(Icons.wechat, color: Colors.green),
              title: const Text('分享到微信'),
              onTap: () {
                Navigator.pop(ctx);
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请使用系统分享')));
              },
            ),
          ]),
        ),
      ),
    );
  }

  String _getStatusText() {
    final now = DateTime.now();
    if (event.endTime.isBefore(now)) return '已结束';
    if (event.startTime.isAfter(now)) return '即将开始';
    return '进行中';
  }

  String _formatDate(DateTime dt) => '${dt.year}年${dt.month}月${dt.day}日';
  String _formatTime(DateTime dt) => '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
}

class _RegisterButton extends StatefulWidget {
  final EventModel event;
  const _RegisterButton({required this.event});
  @override
  State<_RegisterButton> createState() => _RegisterButtonState();
}

class _RegisterButtonState extends State<_RegisterButton> {
  bool _registered = false;
  bool _loading = false;

  Future<void> _register() async {
    setState(() => _loading = true);
    try {
      final res = await ApiClient().post('/events/${widget.event.id}/register');
      final d = res.data;
      if (d is Map && d['code'] == 0) {
        if (mounted) {
          setState(() => _registered = true);
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('报名成功！')));
        }
      } else {
        final msg = d is Map ? (d['message'] ?? '报名失败') : '报名失败';
        if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg.toString())));
      }
    } catch (_) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('报名失败，请重试')));
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity, height: 48,
      child: _registered
          ? FilledButton.icon(
              onPressed: null,
              icon: const Icon(Icons.check_circle),
              label: const Text('已报名', style: TextStyle(fontSize: 16)),
              style: FilledButton.styleFrom(backgroundColor: Colors.green),
            )
          : FilledButton.icon(
              onPressed: _loading ? null : _register,
              icon: _loading
                  ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                  : const Icon(Icons.check_circle_outline),
              label: const Text('我要报名', style: TextStyle(fontSize: 16)),
            ),
    );
  }
}

class _InfoCard extends StatelessWidget {
  final IconData icon; final String label; final String value;
  const _InfoCard({required this.icon, required this.label, required this.value});
  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: 40, height: 40,
          decoration: BoxDecoration(
            color: Theme.of(context).primaryColor.withOpacity(0.1),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Icon(icon, size: 20, color: Theme.of(context).primaryColor),
        ),
        const SizedBox(width: 12),
        Expanded(child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(label, style: const TextStyle(fontSize: 12, color: Colors.grey)),
            const SizedBox(height: 2),
            Text(value, style: const TextStyle(fontSize: 15)),
          ],
        )),
      ],
    );
  }
}
