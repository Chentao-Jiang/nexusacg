import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/network/api_client.dart';

class MyProductsScreen extends StatefulWidget {
  const MyProductsScreen({super.key});

  @override
  State<MyProductsScreen> createState() => _MyProductsScreenState();
}

class _MyProductsScreenState extends State<MyProductsScreen> {
  List<Map<String, dynamic>> _items = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final res = await ApiClient().get('/products/my');
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('我的商品')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _items.isEmpty
              ? const Center(child: Text('暂无商品'))
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
                        title: Text(item['name'] ?? '', maxLines: 1, overflow: TextOverflow.ellipsis),
                        subtitle: Row(
                          children: [
                            Text('¥${item['price'] ?? 0}'),
                            const SizedBox(width: 12),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                              decoration: BoxDecoration(
                                color: (item['status'] == 'active') ? Colors.green.shade50 : Colors.grey.shade100,
                                borderRadius: BorderRadius.circular(3),
                              ),
                              child: Text(
                                item['status'] == 'active' ? '在售' : '下架',
                                style: TextStyle(
                                  fontSize: 10,
                                  color: item['status'] == 'active' ? Colors.green : Colors.grey,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
    );
  }
}
