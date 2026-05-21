import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/presentation/screens/groups/group_detail_screen.dart';
import 'package:nexusacg/presentation/screens/groups/group_create_screen.dart';

class GroupListScreen extends StatefulWidget {
  const GroupListScreen({super.key});
  @override
  State<GroupListScreen> createState() => _GroupListScreenState();
}

class _GroupListScreenState extends State<GroupListScreen> {
  final _api = ApiClient();
  List<Map<String, dynamic>> _hotGroups = [];
  List<Map<String, dynamic>> _myGroups = [];
  bool _loading = true;

  @override
  void initState() { super.initState(); _load(); }

  Future<void> _load() async {
    setState(() => _loading = true);
    final hotRes = await _api.get('/groups', queryParameters: {'sort': 'popular', 'page_size': '10'});
    final myRes = await _api.get('/groups/my');
    if (mounted) {
      setState(() {
        _hotGroups = _parseItems(hotRes.data);
        _myGroups = _parseItems(myRes.data);
        _loading = false;
      });
    }
  }

  List<Map<String, dynamic>> _parseItems(dynamic data) {
    if (data is Map && data['code'] == 0 && data['data'] != null) {
      final d = data['data'] as Map;
      return (d['items'] as List?)?.cast<Map<String, dynamic>>() ?? [];
    }
    return [];
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('兴趣圈层'),
        actions: [
          IconButton(icon: const Icon(Icons.search), onPressed: () {}),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _load,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  if (_myGroups.isNotEmpty) ...[
                    const Text('我的小组', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                    const SizedBox(height: 12),
                    SizedBox(
                      height: 150,
                      child: ListView.builder(
                        scrollDirection: Axis.horizontal,
                        itemCount: _myGroups.length,
                        itemBuilder: (_, i) => _groupCard(_myGroups[i]),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ],
                  const Text('热门小组', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 12),
                  ..._hotGroups.map((g) => _groupTile(g)),
                ],
              ),
            ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const GroupCreateScreen())).then((_) => _load()),
        child: const Icon(Icons.add),
      ),
    );
  }

  Widget _groupCard(Map<String, dynamic> g) {
    final cover = g['cover_url'] as String?;
    return GestureDetector(
      onTap: () => _openGroup(g),
      child: Container(
        width: 120,
        margin: const EdgeInsets.only(right: 12),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          color: Colors.white,
          boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 8)],
        ),
        child: Column(
          children: [
            ClipRRect(
              borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
              child: cover != null
                  ? CachedNetworkImage(imageUrl: cover, height: 80, width: 120, fit: BoxFit.cover)
                  : Container(height: 80, color: Colors.grey.shade300),
            ),
            Padding(
              padding: const EdgeInsets.all(8),
              child: Column(
                children: [
                  Text(g['name'] ?? '', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w500, fontSize: 13)),
                  Text('${g['member_count'] ?? 0} 成员', style: const TextStyle(fontSize: 11, color: Colors.grey)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _groupTile(Map<String, dynamic> g) {
    final cover = g['cover_url'] as String?;
    return Card(
      margin: const EdgeInsets.only(bottom: 10),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => _openGroup(g),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(10),
                child: cover != null
                    ? CachedNetworkImage(imageUrl: cover, width: 60, height: 60, fit: BoxFit.cover)
                    : Container(width: 60, height: 60, color: Colors.grey.shade300, child: const Icon(Icons.group)),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(g['name'] ?? '', style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 15)),
                    const SizedBox(height: 4),
                    Text(g['description'] ?? '', maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(color: Colors.grey, fontSize: 13)),
                    const SizedBox(height: 6),
                    Text('${g['member_count'] ?? 0} 成员', style: const TextStyle(fontSize: 12, color: Colors.grey)),
                  ],
                ),
              ),
              const Icon(Icons.chevron_right, color: Colors.grey),
            ],
          ),
        ),
      ),
    );
  }

  void _openGroup(Map<String, dynamic> g) {
    Navigator.push(context, MaterialPageRoute(
      builder: (_) => GroupDetailScreen(groupId: g['id']?.toString() ?? ''),
    )).then((_) => _load());
  }
}
