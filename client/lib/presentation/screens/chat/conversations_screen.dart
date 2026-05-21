import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/presentation/screens/chat/chat_screen.dart';

class ConversationsScreen extends StatefulWidget {
  const ConversationsScreen({super.key});
  @override
  State<ConversationsScreen> createState() => _ConversationsScreenState();
}

class _ConversationsScreenState extends State<ConversationsScreen> {
  List<Map<String, dynamic>> _convs = [];
  bool _loading = true;

  @override
  void initState() { super.initState(); _load(); }

  Future<void> _load() async {
    final res = await ApiClient().get('/messages/conversations');
    final d = res.data;
    if (d is Map && d['code'] == 0 && d['data'] != null) {
      final data = d['data'] as Map;
      setState(() {
        _convs = (data['items'] as List?)?.cast<Map<String, dynamic>>() ?? [];
        _loading = false;
      });
    } else {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('消息')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _convs.isEmpty
              ? const Center(child: Text('暂无消息'))
              : ListView.separated(
                  itemCount: _convs.length,
                  separatorBuilder: (_, __) => const Divider(height: 1, indent: 72),
                  itemBuilder: (_, i) {
                    final c = _convs[i];
                    final unread = (c['unread'] as num?)?.toInt() ?? 0;
                    return ListTile(
                      leading: CircleAvatar(
                        radius: 24,
                        backgroundImage: c['other_avatar_url'] != null
                            ? CachedNetworkImageProvider(c['other_avatar_url'].toString())
                            : null,
                      ),
                      title: Row(children: [
                        Text(c['other_nickname'] ?? '用户', style: const TextStyle(fontWeight: FontWeight.w500)),
                        if (unread > 0) ...[
                          const SizedBox(width: 8),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                            decoration: BoxDecoration(color: Colors.red, borderRadius: BorderRadius.circular(10)),
                            child: Text('$unread', style: const TextStyle(color: Colors.white, fontSize: 11)),
                          ),
                        ],
                      ]),
                      subtitle: Text(c['last_message'] ?? '', maxLines: 1, overflow: TextOverflow.ellipsis),
                      onTap: () {
                        Navigator.push(context, MaterialPageRoute(
                          builder: (_) => ChatScreen(
                            otherUserId: c['other_user_id'].toString(),
                            otherName: c['other_nickname'] ?? '用户',
                          ),
                        )).then((_) => _load());
                      },
                    );
                  },
                ),
    );
  }
}
