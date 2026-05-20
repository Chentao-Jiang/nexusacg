import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/repositories/follow_repository.dart';

class FollowListScreen extends StatefulWidget {
  final String userId;
  final String title; // '粉丝' or '关注'
  final bool isFollowers; // true = followers, false = following
  const FollowListScreen({super.key, required this.userId, required this.title, required this.isFollowers});

  @override
  State<FollowListScreen> createState() => _FollowListScreenState();
}

class _FollowListScreenState extends State<FollowListScreen> {
  final _repo = FollowRepository();
  List<Map<String, dynamic>> _users = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final result = widget.isFollowers
        ? await _repo.getFollowers(widget.userId)
        : await _repo.getFollowing(widget.userId);
    if (mounted && result != null) {
      final items = result['items'] as List? ?? [];
      setState(() {
        _users = items.cast<Map<String, dynamic>>();
        _loading = false;
      });
    } else if (mounted) {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(widget.title)),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _users.isEmpty
              ? Center(child: Text('暂无${widget.title}'))
              : ListView.separated(
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  itemCount: _users.length,
                  separatorBuilder: (_, __) => const Divider(height: 1, indent: 72),
                  itemBuilder: (ctx, i) {
                    final u = _users[i];
                    final avatar = u['avatar_url'] as String?;
                    final nickname = u['nickname'] as String? ?? '用户';
                    return ListTile(
                      leading: CircleAvatar(
                        radius: 20,
                        backgroundImage: avatar != null ? CachedNetworkImageProvider(avatar) : null,
                        child: avatar == null ? const Icon(Icons.person) : null,
                      ),
                      title: Text(nickname, style: const TextStyle(fontWeight: FontWeight.w500)),
                    );
                  },
                ),
    );
  }
}
