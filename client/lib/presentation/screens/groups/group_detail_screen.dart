import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/presentation/screens/community/post_detail_screen.dart';
import 'package:nexusacg/presentation/screens/community/post_create_screen.dart';
import 'package:nexusacg/core/models/models.dart';

class GroupDetailScreen extends StatefulWidget {
  final String groupId;
  const GroupDetailScreen({super.key, required this.groupId});
  @override
  State<GroupDetailScreen> createState() => _GroupDetailScreenState();
}

class _GroupDetailScreenState extends State<GroupDetailScreen> with SingleTickerProviderStateMixin {
  final _api = ApiClient();
  Map<String, dynamic>? _group;
  List<Map<String, dynamic>> _posts = [];
  List<Map<String, dynamic>> _members = [];
  bool _loading = true;
  bool _joined = false;
  late TabController _tabCtrl;

  @override
  void initState() { super.initState(); _tabCtrl = TabController(length: 2, vsync: this); _load(); }

  Future<void> _load() async {
    final gRes = await _api.get('/groups/${widget.groupId}');
    final pRes = await _api.get('/groups/${widget.groupId}/posts');
    final mRes = await _api.get('/groups/${widget.groupId}/members');

    if (mounted) {
      if (gRes.data is Map && gRes.data['code'] == 0) {
        setState(() => _group = (gRes.data as Map)['data'] as Map<String, dynamic>?);
      }
      setState(() {
        _posts = _parseList(pRes.data, 'posts');
        _members = _parseList(mRes.data, 'members');
        _loading = false;
      });
      _checkIfJoined();
    }
  }

  List<Map<String, dynamic>> _parseList(dynamic data, String sourceType) {
    if (data is Map && data['code'] == 0 && data['data'] != null) {
      final d = data['data'] as Map;
      final items = d['items'] as List?;
      // Handle both raw maps and GroupMember objects
      return items?.map((e) {
        if (e is Map<String, dynamic>) return e;
        return <String, dynamic>{};
      }).toList() ?? [];
    }
    return [];
  }

  Future<void> _checkIfJoined() async {
    final res = await _api.get('/groups/my');
    final data = _parseList(res.data, 'groups');
    final joined = data.any((g) => g['id']?.toString() == widget.groupId);
    if (mounted) setState(() => _joined = joined);
  }

  Future<void> _join() async {
    await _api.post('/groups/${widget.groupId}/join');
    setState(() => _joined = true);
    _load();
  }

  Future<void> _leave() async {
    await _api.post('/groups/${widget.groupId}/leave');
    setState(() => _joined = false);
    _load();
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) return Scaffold(appBar: AppBar(title: const Text('小组')), body: const Center(child: CircularProgressIndicator()));
    final g = _group;
    if (g == null) return Scaffold(appBar: AppBar(title: const Text('小组')), body: const Center(child: Text('小组不存在')));
    final cover = g['cover_url'] as String?;

    return Scaffold(
      body: NestedScrollView(
        headerSliverBuilder: (_, __) => [
          SliverAppBar(
            expandedHeight: 180,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              background: cover != null
                  ? CachedNetworkImage(imageUrl: cover, fit: BoxFit.cover)
                  : Container(color: Colors.grey.shade300),
            ),
            actions: [
              if (_joined)
                TextButton(onPressed: _leave, child: const Text('退出', style: TextStyle(color: Colors.white))),
              if (!_joined)
                TextButton(onPressed: _join, child: const Text('加入', style: TextStyle(color: Colors.white))),
            ],
          ),
        ],
        body: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(g['name'] ?? '', style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 4),
                  Text('${g['member_count'] ?? 0} 成员', style: const TextStyle(color: Colors.grey)),
                  const SizedBox(height: 8),
                  Text(g['description'] ?? '', style: const TextStyle(fontSize: 14, height: 1.5)),
                ],
              ),
            ),
            TabBar(
              controller: _tabCtrl,
              tabs: const [Tab(text: '帖子'), Tab(text: '成员')],
            ),
            Expanded(
              child: TabBarView(
                controller: _tabCtrl,
                children: [
                  _buildPostsTab(),
                  _buildMembersTab(),
                ],
              ),
            ),
          ],
        ),
      ),
      floatingActionButton: _joined ? FloatingActionButton(
        onPressed: () => Navigator.push(context, MaterialPageRoute(
          builder: (_) => PostCreateScreen(groupId: widget.groupId),
        )).then((_) => _load()),
        child: const Icon(Icons.edit),
      ) : null,
    );
  }

  Widget _buildPostsTab() {
    if (_posts.isEmpty) return const Center(child: Text('暂无帖子'));
    return ListView.builder(
      padding: const EdgeInsets.all(12),
      itemCount: _posts.length,
      itemBuilder: (_, i) {
        final p = _posts[i];
        return Card(
          margin: const EdgeInsets.only(bottom: 8),
          child: ListTile(
            leading: CircleAvatar(
              radius: 18,
              backgroundImage: p['author']?['avatar_url'] != null
                  ? CachedNetworkImageProvider(p['author']['avatar_url'])
                  : null,
            ),
            title: Text(p['title']?.toString() ?? '', maxLines: 1, overflow: TextOverflow.ellipsis),
            subtitle: Text(p['content']?.toString() ?? '', maxLines: 2, overflow: TextOverflow.ellipsis),
            onTap: () {
              try {
                final post = PostModel.fromJson(p);
                Navigator.push(context, MaterialPageRoute(builder: (_) => PostDetailScreen(post: post)));
              } catch (_) {}
            },
          ),
        );
      },
    );
  }

  Widget _buildMembersTab() {
    if (_members.isEmpty) return const Center(child: Text('暂无成员'));
    return ListView.builder(
      padding: const EdgeInsets.all(12),
      itemCount: _members.length,
      itemBuilder: (_, i) {
        final m = _members[i];
        final user = m['user'] as Map<String, dynamic>?;
        final avatar = user?['avatar_url'] as String?;
        final nickname = user?['nickname'] ?? '用户';
        final role = m['role'] ?? 'member';
        return ListTile(
          leading: CircleAvatar(
            radius: 18,
            backgroundImage: avatar != null ? CachedNetworkImageProvider(avatar) : null,
            child: avatar == null ? const Icon(Icons.person, size: 18) : null,
          ),
          title: Text(nickname),
          trailing: role == 'owner' ? Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
            decoration: BoxDecoration(color: Colors.orange.shade100, borderRadius: BorderRadius.circular(10)),
            child: const Text('组长', style: TextStyle(color: Colors.orange, fontSize: 11)),
          ) : null,
        );
      },
    );
  }

  @override
  void dispose() { _tabCtrl.dispose(); super.dispose(); }
}
