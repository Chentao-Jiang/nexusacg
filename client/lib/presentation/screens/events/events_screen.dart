import 'package:flutter/material.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/screens/events/event_create_screen.dart';
import 'package:nexusacg/presentation/screens/events/event_detail_screen.dart';

class EventsScreen extends StatefulWidget {
  const EventsScreen({super.key});

  @override
  State<EventsScreen> createState() => _EventsScreenState();
}

class _EventsScreenState extends State<EventsScreen> {
  final _repo = EventRepository();
  List<EventModel> _events = [];
  bool _loading = true;
  int _page = 1;
  bool _hasMore = true;

  @override
  void initState() {
    super.initState();
    _loadEvents();
  }

  Future<void> _loadEvents() async {
    setState(() => _loading = true);
    try {
      final result = await _repo.getEvents(page: _page, status: 'upcoming');
      setState(() {
        _events = result.items;
        _hasMore = result.items.length >= 20;
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加载失败: $e')),
        );
      }
    }
  }

  Future<void> _loadMore() async {
    if (!_hasMore) return;
    final nextPage = _page + 1;
    try {
      final result = await _repo.getEvents(page: nextPage, status: 'upcoming');
      setState(() {
        _events.addAll(result.items);
        _page = nextPage;
        _hasMore = result.items.length >= 20;
      });
    } catch (e) {
      // Ignore load more errors
    }
  }

  void _goToDetail(EventModel event) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => EventDetailScreen(event: event)),
    );
  }

  void _goToCreate() {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => const EventCreateScreen()),
    ).then((_) => _loadEvents());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('活动'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: _goToCreate,
            tooltip: '创建活动',
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: () async => _loadEvents(),
              child: _events.isEmpty
                  ? _buildEmptyState()
                  : NotificationListener<ScrollNotification>(
                      onNotification: (notification) {
                        if (notification is ScrollEndNotification &&
                            notification.metrics.pixels >=
                                notification.metrics.maxScrollExtent * 0.8) {
                          _loadMore();
                        }
                        return false;
                      },
                      child: ListView.builder(
                        padding: const EdgeInsets.all(12),
                        itemCount: _events.length,
                        itemBuilder: (context, index) => GestureDetector(
                          onTap: () => _goToDetail(_events[index]),
                          child: _EventCard(_events[index]),
                        ),
                      ),
                    ),
            ),
    );
  }

  Widget _buildEmptyState() {
    return ListView(
      children: [
        Container(
          height: 200,
          margin: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: Colors.grey.shade100,
            borderRadius: BorderRadius.circular(16),
          ),
          child: const Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.event_busy, size: 48, color: Colors.grey),
              SizedBox(height: 12),
              Text('暂无活动', style: TextStyle(color: Colors.grey)),
              SizedBox(height: 4),
              Text('点击右上角创建活动', style: TextStyle(color: Colors.grey, fontSize: 12)),
            ],
          ),
        ),
      ],
    );
  }
}

class _EventCard extends StatelessWidget {
  final EventModel event;
  const _EventCard(this.event);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isPast = event.endTime.isBefore(DateTime.now());
    final statusText = _getStatusText();

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Date badge
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                color: theme.primaryColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    '${event.startTime.month}月',
                    style: TextStyle(
                      fontSize: 11,
                      color: theme.primaryColor,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  Text(
                    '${event.startTime.day}',
                    style: TextStyle(
                      fontSize: 20,
                      color: theme.primaryColor,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            // Event info
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          event.name,
                          style: const TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 16,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      if (isPast)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.grey.shade200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            statusText,
                            style: const TextStyle(
                              fontSize: 10,
                              color: Colors.grey,
                            ),
                          ),
                        ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Row(
                    children: [
                      const Icon(Icons.location_on_outlined, size: 14, color: Colors.grey),
                      const SizedBox(width: 4),
                      Expanded(
                        child: Text(
                          event.address,
                          style: const TextStyle(color: Colors.grey, fontSize: 13),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      const Icon(Icons.access_time_outlined, size: 14, color: Colors.grey),
                      const SizedBox(width: 4),
                      Text(
                        '${_formatTime(event.startTime)} - ${_formatTime(event.endTime)}',
                        style: const TextStyle(color: Colors.grey, fontSize: 13),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const Icon(Icons.chevron_right, color: Colors.grey),
          ],
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

  String _formatTime(DateTime dt) {
    return '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
  }
}
