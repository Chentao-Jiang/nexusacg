import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/network/api_client.dart';

class ProviderDetailScreen extends StatefulWidget {
  final String providerId;
  const ProviderDetailScreen({super.key, required this.providerId});
  @override
  State<ProviderDetailScreen> createState() => _ProviderDetailScreenState();
}

class _ProviderDetailScreenState extends State<ProviderDetailScreen> with SingleTickerProviderStateMixin {
  Map<String, dynamic>? _sp;
  List<Map<String, dynamic>> _reviews = [];
  bool _loading = true;
  final _bookingNotes = TextEditingController();
  late TabController _tabCtrl;

  @override
  void initState() { super.initState(); _tabCtrl = TabController(length: 2, vsync: this); _load(); }

  Future<void> _load() async {
    final spRes = await ApiClient().get('/service-providers/${widget.providerId}');
    final rvRes = await ApiClient().get('/service-providers/${widget.providerId}/reviews');
    if (mounted) {
      if (spRes.data is Map && spRes.data['code'] == 0) {
        setState(() => _sp = (spRes.data as Map)['data'] as Map<String, dynamic>?);
      }
      setState(() {
        _reviews = _parseList(rvRes.data);
        _loading = false;
      });
    }
  }

  List<Map<String, dynamic>> _parseList(dynamic data) {
    if (data is Map && data['code'] == 0 && data['data'] != null) {
      final d = data['data'] as Map;
      return (d['items'] as List?)?.cast<Map<String, dynamic>>() ?? [];
    }
    return [];
  }

  Future<void> _book() async {
    await showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('预约服务'),
        content: TextField(controller: _bookingNotes, decoration: const InputDecoration(hintText: '备注（可选，如时间要求）'), maxLines: 3),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          FilledButton(onPressed: () async {
            Navigator.pop(ctx);
            await ApiClient().post('/service-providers/${widget.providerId}/book', data: {'notes': _bookingNotes.text, 'service_type': _sp?['type'] ?? ''});
            _bookingNotes.clear();
            if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('预约请求已发送')));
          }, child: const Text('确认预约')),
        ],
      ),
    );
  }

  Future<void> _addReview() async {
    final ratingCtrl = TextEditingController();
    final commentCtrl = TextEditingController();
    int rating = 5;
    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          title: const Text('评价服务者'),
          content: Column(mainAxisSize: MainAxisSize.min, children: [
            Row(mainAxisAlignment: MainAxisAlignment.center, children: List.generate(5, (i) => IconButton(
              icon: Icon(i < rating ? Icons.star : Icons.star_border, color: Colors.amber, size: 32),
              onPressed: () => setDialogState(() => rating = i + 1),
            ))),
            TextField(controller: commentCtrl, decoration: const InputDecoration(hintText: '写下你的评价'), maxLines: 3),
          ]),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
            FilledButton(onPressed: () async {
              Navigator.pop(ctx);
              await ApiClient().post('/service-providers/${widget.providerId}/review', data: {'rating': rating, 'comment': commentCtrl.text});
              _load();
            }, child: const Text('提交')),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) return Scaffold(appBar: AppBar(title: const Text('服务者详情')), body: const Center(child: CircularProgressIndicator()));
    final sp = _sp;
    if (sp == null) return Scaffold(appBar: AppBar(title: const Text('服务者详情')), body: const Center(child: Text('未找到')));

    final images = sp['portfolio_images'] as List?;
    final cover = images != null && images.isNotEmpty ? images[0].toString() : null;

    return Scaffold(
      body: NestedScrollView(
        headerSliverBuilder: (_, __) => [
          SliverAppBar(
            expandedHeight: 200, pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              background: cover != null ? CachedNetworkImage(imageUrl: cover, fit: BoxFit.cover) : Container(color: Colors.grey.shade300),
            ),
          ),
        ],
        body: Column(children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Row(children: [
                Expanded(child: Text(_typeLabel(sp['type']), style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold))),
                if (sp['is_verified'] == true) const Icon(Icons.verified, color: Colors.blue, size: 24),
              ]),
              const SizedBox(height: 8),
              Row(children: [
                const Icon(Icons.star, size: 16, color: Colors.amber),
                Text(' ${sp['rating']?.toString() ?? "0"}', style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500)),
                Text(' (${sp['review_count'] ?? 0} 评价)', style: const TextStyle(fontSize: 14, color: Colors.grey)),
              ]),
              const SizedBox(height: 8),
              Text(sp['description']?.toString() ?? '', style: const TextStyle(fontSize: 14, height: 1.6)),
              if (sp['price_list'] is List && (sp['price_list'] as List).isNotEmpty) ...[
                const SizedBox(height: 12),
                const Text('价格参考', style: TextStyle(fontWeight: FontWeight.w600)),
                ...((sp['price_list'] as List).map((p) => Padding(
                  padding: const EdgeInsets.only(top: 4),
                  child: Text('${p}', style: const TextStyle(color: Colors.grey, fontSize: 13)),
                ))),
              ],
            ]),
          ),
          TabBar(controller: _tabCtrl, tabs: const [Tab(text: '作品'), Tab(text: '评价')]),
          Expanded(
            child: TabBarView(controller: _tabCtrl, children: [
              images != null && images.isNotEmpty
                  ? GridView.builder(padding: const EdgeInsets.all(8), gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(crossAxisCount: 3, crossAxisSpacing: 4, mainAxisSpacing: 4),
                      itemCount: images.length, itemBuilder: (_, i) => ClipRRect(borderRadius: BorderRadius.circular(8), child: CachedNetworkImage(imageUrl: images[i].toString(), fit: BoxFit.cover)))
                  : const Center(child: Text('暂无作品')),
              _reviews.isEmpty
                  ? const Center(child: Text('暂无评价'))
                  : ListView.builder(padding: const EdgeInsets.all(12), itemCount: _reviews.length, itemBuilder: (_, i) {
                      final r = _reviews[i];
                      return Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: Padding(
                          padding: const EdgeInsets.all(12),
                          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                            Row(children: [
                              CircleAvatar(radius: 14, backgroundImage: r['avatar_url'] != null ? CachedNetworkImageProvider(r['avatar_url']) : null),
                              const SizedBox(width: 8),
                              Text(r['user_name'] ?? '匿名', style: const TextStyle(fontWeight: FontWeight.w500, fontSize: 13)),
                              const Spacer(),
                              Row(children: List.generate((r['rating'] as num?)?.toInt() ?? 5, (_) => const Icon(Icons.star, size: 14, color: Colors.amber))),
                            ]),
                            if ((r['comment'] as String?)?.isNotEmpty == true) ...[
                              const SizedBox(height: 6),
                              Text(r['comment'] ?? '', style: const TextStyle(fontSize: 14)),
                            ],
                          ]),
                        ),
                      );
                    }),
            ]),
          ),
        ]),
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(children: [
            Expanded(
              child: FilledButton.icon(onPressed: _book, icon: const Icon(Icons.calendar_today), label: const Text('预约服务')),
            ),
            const SizedBox(width: 12),
            OutlinedButton.icon(onPressed: _addReview, icon: const Icon(Icons.star_border), label: const Text('评价')),
          ]),
        ),
      ),
    );
  }

  String _typeLabel(String? t) {
    return {'makeup_artist':'妆娘','wig_stylist':'毛娘','photographer':'摄影','post_editor':'后期','props_maker':'道具'}[t ?? ''] ?? t ?? '';
  }

  @override
  void dispose() { _tabCtrl.dispose(); _bookingNotes.dispose(); super.dispose(); }
}
