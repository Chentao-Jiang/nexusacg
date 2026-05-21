import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/presentation/screens/services/provider_detail_screen.dart';

class ProviderListScreen extends StatefulWidget {
  const ProviderListScreen({super.key});
  @override
  State<ProviderListScreen> createState() => _ProviderListScreenState();
}

class _ProviderListScreenState extends State<ProviderListScreen> {
  List<Map<String, dynamic>> _items = [];
  bool _loading = true;
  String _filter = '';

  @override
  void initState() { super.initState(); _load(); }

  Future<void> _load() async {
    final params = <String, dynamic>{'page_size': '50'};
    if (_filter.isNotEmpty) params['type'] = _filter;
    final res = await ApiClient().get('/service-providers', queryParameters: params);
    if (mounted) {
      final d = res.data;
      if (d is Map && d['code'] == 0) {
        final data = d['data'] as Map;
        setState(() { _items = (data['items'] as List?)?.cast<Map<String, dynamic>>() ?? []; _loading = false; });
      } else { setState(() => _loading = false); }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('服务者')),
      body: Column(
        children: [
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.all(12),
            child: Row(
              children: ['', 'makeup_artist', 'wig_stylist', 'photographer', 'post_editor', 'props_maker'].map((t) {
                final label = t.isEmpty ? '全部' : {'makeup_artist':'妆娘','wig_stylist':'毛娘','photographer':'摄影','post_editor':'后期','props_maker':'道具'}[t] ?? t;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: ChoiceChip(
                    label: Text(label), selected: _filter == t,
                    onSelected: (_) { setState(() { _filter = t; _loading = true; }); _load(); },
                  ),
                );
              }).toList(),
            ),
          ),
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : _items.isEmpty
                    ? const Center(child: Text('暂无服务者'))
                    : ListView.builder(
                        padding: const EdgeInsets.symmetric(horizontal: 12),
                        itemCount: _items.length,
                        itemBuilder: (_, i) {
                          final sp = _items[i];
                          final images = sp['portfolio_images'] as List?;
                          final cover = images != null && images.isNotEmpty ? images[0].toString() : null;
                          return Card(
                            margin: const EdgeInsets.only(bottom: 10),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                            child: InkWell(
                              borderRadius: BorderRadius.circular(12),
                              onTap: () => Navigator.push(context, MaterialPageRoute(
                                builder: (_) => ProviderDetailScreen(providerId: sp['id']?.toString() ?? ''),
                              )),
                              child: Padding(
                                padding: const EdgeInsets.all(12),
                                child: Row(
                                  children: [
                                    ClipRRect(
                                      borderRadius: BorderRadius.circular(10),
                                      child: cover != null
                                          ? CachedNetworkImage(imageUrl: cover, width: 72, height: 72, fit: BoxFit.cover)
                                          : Container(width: 72, height: 72, color: Colors.grey.shade300, child: const Icon(Icons.person, size: 32)),
                                    ),
                                    const SizedBox(width: 12),
                                    Expanded(
                                      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                                        Row(children: [
                                          Text(_typeLabel(sp['type']), style: const TextStyle(fontSize: 12, color: Colors.blue)),
                                          if (sp['is_verified'] == true) ...[
                                            const SizedBox(width: 4),
                                            const Icon(Icons.verified, size: 14, color: Colors.blue),
                                          ],
                                        ]),
                                        const SizedBox(height: 4),
                                        Text(sp['description']?.toString() ?? '', maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 14)),
                                        const SizedBox(height: 6),
                                        Row(children: [
                                          const Icon(Icons.star, size: 14, color: Colors.amber),
                                          Text(' ${sp['rating']?.toString() ?? "0"}', style: const TextStyle(fontSize: 12)),
                                          const SizedBox(width: 12),
                                          Text('${sp['review_count'] ?? 0} 评价', style: const TextStyle(fontSize: 12, color: Colors.grey)),
                                        ]),
                                      ]),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }

  String _typeLabel(String? t) {
    return {'makeup_artist':'妆娘','wig_stylist':'毛娘','photographer':'摄影','post_editor':'后期','props_maker':'道具'}[t ?? ''] ?? t ?? '';
  }
}
