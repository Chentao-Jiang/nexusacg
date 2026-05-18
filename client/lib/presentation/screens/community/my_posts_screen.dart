import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/network/api_client.dart';

class MyPostsScreen extends StatefulWidget {
  const MyPostsScreen({super.key});

  @override
  State<MyPostsScreen> createState() => _MyPostsScreenState();
}

class _MyPostsScreenState extends State<MyPostsScreen> {
  List<PostModel> _posts = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadMyPosts();
  }

  Future<void> _loadMyPosts() async {
    setState(() => _loading = true);
    try {
      final response = await ApiClient().get('/posts/my');
      final data = response.data;
      if (data is Map && data['code'] == 0 && data['data'] != null) {
        final items = data['data']['items'] as List? ?? [];
        setState(() {
          _posts = items.map((e) => PostModel.fromJson(e as Map<String, dynamic>)).toList();
          _loading = false;
        });
      } else {
        setState(() => _loading = false);
      }
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  Future<void> _deletePost(String postId) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('删除帖子'),
        content: const Text('确定要删除这条帖子吗？此操作不可撤销。'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('取消')),
          TextButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('删除', style: TextStyle(color: Colors.red))),
        ],
      ),
    );
    if (confirm != true) return;
    try {
      await ApiClient().delete('/posts/$postId');
      _loadMyPosts();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('已删除')));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('删除失败: $e')));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('我的帖子')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _posts.isEmpty
              ? const Center(child: Text('暂无帖子'))
              : RefreshIndicator(
                  onRefresh: _loadMyPosts,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(12),
                    itemCount: _posts.length,
                    itemBuilder: (context, index) {
                      final post = _posts[index];
                      return Card(
                        margin: const EdgeInsets.only(bottom: 12),
                        child: ListTile(
                          leading: post.images.isNotEmpty
                              ? ClipRRect(
                                  borderRadius: BorderRadius.circular(4),
                                  child: CachedNetworkImage(
                                    imageUrl: post.images.first,
                                    width: 56, height: 56, fit: BoxFit.cover,
                                    errorWidget: (_, __, ___) => Container(width: 56, height: 56, color: Colors.grey.shade200),
                                  ),
                                )
                              : null,
                          title: Text(post.title.isNotEmpty ? post.title : post.content, maxLines: 2, overflow: TextOverflow.ellipsis),
                          subtitle: Text('${post.likeCount} 赞  ${post.commentCount} 评论  ${post.status}',
                              style: const TextStyle(fontSize: 12, color: Colors.grey)),
                          trailing: PopupMenuButton<String>(
                            onSelected: (action) {
                              if (action == 'delete') _deletePost(post.id);
                            },
                            itemBuilder: (_) => [
                              const PopupMenuItem(value: 'delete', child: Text('删除', style: TextStyle(color: Colors.red))),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
    );
  }
}
