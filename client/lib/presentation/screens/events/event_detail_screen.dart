import 'package:flutter/material.dart';
import 'package:nexusacg/core/models/models.dart';

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
          // Hero image / cover
          SliverAppBar(
            expandedHeight: 200,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              title: Text(event.name, style: const TextStyle(fontSize: 16)),
              background: event.coverUrl != null
                  ? Image.network(
                      event.coverUrl!,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => _buildDefaultCover(theme),
                    )
                  : _buildDefaultCover(theme),
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.share),
                onPressed: () {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('分享功能开发中')),
                  );
                },
              ),
            ],
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Status badge
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 10,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: isPast ? Colors.grey.shade200 : theme.primaryColor.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Text(
                          statusText,
                          style: TextStyle(
                            color: isPast ? Colors.grey : theme.primaryColor,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),

                  // Name
                  Text(
                    event.name,
                    style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 20),

                  // Info cards
                  _InfoCard(
                    icon: Icons.calendar_today,
                    label: '时间',
                    value: '${_formatDate(event.startTime)} ${_formatTime(event.startTime)} - ${_formatTime(event.endTime)}',
                  ),
                  const SizedBox(height: 8),
                  _InfoCard(
                    icon: Icons.location_on,
                    label: '地点',
                    value: event.address,
                  ),

                  const SizedBox(height: 24),
                  const Divider(),
                  const SizedBox(height: 16),

                  // Description
                  const Text(
                    '活动详情',
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    event.description.isEmpty ? '暂无详情' : event.description,
                    style: TextStyle(
                      fontSize: 15,
                      color: event.description.isEmpty ? Colors.grey : null,
                      height: 1.6,
                    ),
                  ),

                  const SizedBox(height: 32),

                  // Action button
                  if (!isPast)
                    SizedBox(
                      width: double.infinity,
                      height: 48,
                      child: FilledButton.icon(
                        onPressed: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('报名功能开发中')),
                          );
                        },
                        icon: const Icon(Icons.check_circle_outline),
                        label: const Text('我要报名', style: TextStyle(fontSize: 16)),
                      ),
                    ),
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
      child: Center(
        child: Icon(Icons.event, size: 64, color: theme.primaryColor.withOpacity(0.5)),
      ),
    );
  }

  String _getStatusText() {
    final now = DateTime.now();
    if (event.endTime.isBefore(now)) return '已结束';
    if (event.startTime.isAfter(now)) return '即将开始';
    return '进行中';
  }

  String _formatDate(DateTime dt) {
    return '${dt.year}年${dt.month}月${dt.day}日';
  }

  String _formatTime(DateTime dt) {
    return '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
  }
}

class _InfoCard extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  const _InfoCard({
    required this.icon,
    required this.label,
    required this.value,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: Theme.of(context).primaryColor.withOpacity(0.1),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Icon(icon, size: 20, color: Theme.of(context).primaryColor),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: const TextStyle(fontSize: 12, color: Colors.grey)),
              const SizedBox(height: 2),
              Text(value, style: const TextStyle(fontSize: 15)),
            ],
          ),
        ),
      ],
    );
  }
}
