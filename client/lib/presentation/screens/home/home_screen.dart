import 'package:flutter/material.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/presentation/screens/products/product_screen.dart';

// Home screen implementation with product zones, featured events, and community feed
class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final _eventRepo = EventRepository();
  List<EventModel> _hotEvents = [];
  bool _eventsLoading = true;

  @override
  void initState() {
    super.initState();
    _loadEvents();
  }

  Future<void> _loadEvents() async {
    try {
      final result = await _eventRepo.getEvents(page: 1, pageSize: 6, status: 'upcoming');
      if (mounted) setState(() {
        _hotEvents = result.items;
        _eventsLoading = false;
      });
    } catch (_) {
      if (mounted) setState(() => _eventsLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('次元链', style: TextStyle(fontWeight: FontWeight.bold)),
        actions: [
          IconButton(icon: const Icon(Icons.search), onPressed: () => _showComingSoon('搜索')),
          IconButton(icon: const Icon(Icons.notifications_outlined), onPressed: () => _showComingSoon('消息通知')),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async { await _loadEvents(); },
        child: CustomScrollView(
          slivers: [
            // Banner
            SliverToBoxAdapter(
              child: Container(
                height: 160,
                margin: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(16),
                  gradient: const LinearGradient(
                    colors: [Color(0xFF6366F1), Color(0xFFEC4899)],
                  ),
                ),
                child: const Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text('欢迎来到次元链', style: TextStyle(fontSize: 24, color: Colors.white, fontWeight: FontWeight.bold)),
                      SizedBox(height: 8),
                      Text('ACG 线下产业生态服务平台', style: TextStyle(fontSize: 14, color: Colors.white70)),
                    ],
                  ),
                ),
              ),
            ),
            // Zone entry
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Row(
                  children: [
                    _zoneCard(context, 'Cosplay专区', Icons.masks, const Color(0xFF6366F1), 0),
                    const SizedBox(width: 12),
                    _zoneCard(context, '周边专区', Icons.shopping_cart, const Color(0xFFEC4899), 1),
                  ],
                ),
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 16)),
            // Hot events
            const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.symmetric(horizontal: 16),
                child: Text('热门活动', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 8)),
            SliverToBoxAdapter(
              child: _eventsLoading
                  ? const SizedBox(height: 120, child: Center(child: CircularProgressIndicator()))
                  : _hotEvents.isEmpty
                      ? const SizedBox(height: 80, child: Center(child: Text('暂无活动', style: TextStyle(color: Colors.grey))))
                      : SizedBox(
                          height: 120,
                          child: ListView.builder(
                            scrollDirection: Axis.horizontal,
                            padding: const EdgeInsets.symmetric(horizontal: 16),
                            itemCount: _hotEvents.length,
                            itemBuilder: (context, index) {
                              final event = _hotEvents[index];
                              return Container(
                                width: 200,
                                margin: const EdgeInsets.only(right: 12),
                                decoration: BoxDecoration(
                                  color: Colors.grey.shade100,
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: Padding(
                                  padding: const EdgeInsets.all(12),
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Text(event.name, maxLines: 2, overflow: TextOverflow.ellipsis,
                                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                                      const SizedBox(height: 4),
                                      Text(_formatDate(event.startTime), maxLines: 1, overflow: TextOverflow.ellipsis,
                                        style: const TextStyle(color: Colors.grey, fontSize: 12)),
                                    ],
                                  ),
                                ),
                              );
                            },
                          ),
                        ),
            ),
            const SliverPadding(padding: EdgeInsets.only(bottom: 24)),
          ],
        ),
      ),
    );
  }

  Widget _zoneCard(BuildContext context, String title, IconData icon, Color color, int tabIndex) {
    return Expanded(
      child: InkWell(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => ProductScreen(initialTabIndex: tabIndex)),
          );
        },
        borderRadius: BorderRadius.circular(16),
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: color.withOpacity(0.1),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Column(
            children: [
              Icon(icon, size: 36, color: color),
              const SizedBox(height: 8),
              Text(title, style: TextStyle(fontWeight: FontWeight.bold, color: color)),
            ],
          ),
        ),
      ),
    );
  }

  void _showComingSoon(String feature) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('$feature 功能开发中')),
    );
  }

  String _formatDate(DateTime dt) {
    return '${dt.month}月${dt.day}日';
  }
}
