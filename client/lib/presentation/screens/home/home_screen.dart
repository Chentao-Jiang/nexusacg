import 'package:flutter/material.dart';
import 'package:nexusacg/presentation/screens/home/home_screen.dart';

// Home screen implementation with product zones, featured events, and community feed
class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('次元链', style: TextStyle(fontWeight: FontWeight.bold)),
        actions: [
          IconButton(icon: const Icon(Icons.search), onPressed: () {}),
          IconButton(icon: const Icon(Icons.notifications_outlined), onPressed: () {}),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {},
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
                    _zoneCard(context, 'Cosplay专区', Icons.costume, const Color(0xFF6366F1)),
                    const SizedBox(width: 12),
                    _zoneCard(context, '周边专区', Icons.shopping_cart, const Color(0xFFEC4899)),
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
              child: SizedBox(
                height: 120,
                child: ListView.builder(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  itemCount: 3,
                  itemBuilder: (context, index) {
                    return Container(
                      width: 200,
                      margin: const EdgeInsets.only(right: 12),
                      decoration: BoxDecoration(
                        color: Colors.grey.shade100,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Center(child: Text('活动占位')),
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

  Widget _zoneCard(BuildContext context, String title, IconData icon, Color color) {
    return Expanded(
      child: InkWell(
        onTap: () {},
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
}
