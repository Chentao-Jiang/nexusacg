import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/presentation/screens/community/post_detail_screen.dart';
import 'package:nexusacg/core/models/models.dart';

class MyBookmarksScreen extends StatefulWidget {
  const MyBookmarksScreen({super.key});

  @override
  State<MyBookmarksScreen> createState() => _MyBookmarksScreenState();
}

class _MyBookmarksScreenState extends State<MyBookmarksScreen> {
  List<Map<String, dynamic>> _items = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final res = await ApiClient().get('/my-bookmarks');
      final d = res.data;
      if (d is Map && d['code'] == 0 && d['data'] != null) {
        final data = d['data'] as Map;
        setState(() {
          _items = (data['items'] as List?)?.cast<Map<String, dynamic>>() ?? [];
          _loading = false;
        });
      } else {
        setState(() => _loading = false);
      }
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  Future<void> _removeBookmark(String postId) async {
    await ApiClient().delete('/posts/$postId/bookmark');
    _load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('我的收藏')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _items.isEmpty
              ? const Center(child: Text('暂无收藏'))
              : ListView.builder(
                  padding: const EdgeInsets.all(12),
                  itemCount: _items.length,
                  itemBuilder: (_, i) {
                    final item = _items[i];
                    final images = item['images'];
                    final imageUrl = images is List && images.isNotEmpty
                        ? images[0].toString()
                        : null;
                    return Card(
                      margin: const EdgeInsets.only(bottom: 10),
                      child: ListTile(
                        leading: ClipRRect(
                          borderRadius: BorderRadius.circular(6),
                          child: imageUrl != null
                              ? CachedNetworkImage(imageUrl: imageUrl, width: 56, height: 56, fit: BoxFit.cover)
                              : Container(width: 56, height: 56, color: Colors.grey.shade200),
                        ),
                        title: Text(item['title']?.toString() ?? item['content']?.toString() ?? '',
                            maxLines: 1, overflow: TextOverflow.ellipsis),
                        subtitle: Text('${item['like_count'] ?? 0} 赞'),
                        trailing: IconButton(
                          icon: const Icon(Icons.bookmark_remove, color: Colors.orange),
                          onPressed: () => _removeBookmark(item['post_id']?.toString() ?? ''),
                        ),
                      ),
                    );
                  },
                ),
    );
  }
}
