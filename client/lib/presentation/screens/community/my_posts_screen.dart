import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/core/repositories/repositories.dart';

class MyPostsScreen extends StatefulWidget {
  const MyPostsScreen({super.key});

  @override
  State<MyPostsScreen> createState() => _MyPostsScreenState();
}

class _MyPostsScreenState extends State<MyPostsScreen> {
  final _repo = PostRepository();
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

  Future<void> _editPost(PostModel post) async {
    final titleCtrl = TextEditingController(text: post.title);
    final contentCtrl = TextEditingController(text: post.content);
    String visibility = post.visibility;

    await showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('编辑帖子'),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: titleCtrl,
                decoration: const InputDecoration(labelText: '标题', border: OutlineInputBorder()),
                maxLength: 200,
              ),
              const SizedBox(height: 12),
              TextField(
                controller: contentCtrl,
                decoration: const InputDecoration(labelText: '内容', border: OutlineInputBorder()),
                maxLines: 5,
                maxLength: 10000,
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  const Icon(Icons.visibility_outlined, size: 20),
                  const SizedBox(width: 8),
                  const Text('可见范围'),
                  const Spacer(),
                  DropdownButton<String>(
                    value: visibility,
                    underline: const SizedBox(),
                    items: const [
                      DropdownMenuItem(value: 'public', child: Text('所有人')),
                      DropdownMenuItem(value: 'followers', child: Text('粉丝可见')),
                      DropdownMenuItem(value: 'private', child: Text('仅自己')),
                    ],
                    onChanged: (v) {
                      if (v != null) {
                        visibility = v;
                        // Trigger rebuild of dropdown
                        (ctx as Element).markNeedsBuild();
                      }
                    },
                  ),
                ],
              ),
            ],
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () async {
              Navigator.pop(ctx);
              try {
                await _repo.updatePost(
                  post.id,
                  title: titleCtrl.text.trim(),
                  content: contentCtrl.text.trim(),
                  visibility: visibility,
                );
                _loadMyPosts();
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('已更新')));
                }
              } catch (e) {
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('更新失败: $e')));
                }
              }
            },
            child: const Text('保存'),
          ),
        ],
      ),
    );
    titleCtrl.dispose();
    contentCtrl.dispose();
  }

  Future<void> _changeVisibility(PostModel post) async {
    final visibility = await showDialog<String>(
      context: context,
      builder: (ctx) => SimpleDialog(
        title: const Text('修改可见范围'),
        children: [
          SimpleDialogOption(
            onPressed: () => Navigator.pop(ctx, 'public'),
            child: const Text('所有人'),
          ),
          SimpleDialogOption(
            onPressed: () => Navigator.pop(ctx, 'followers'),
            child: const Text('粉丝可见'),
          ),
          SimpleDialogOption(
            onPressed: () => Navigator.pop(ctx, 'private'),
            child: const Text('仅自己'),
          ),
        ],
      ),
    );
    if (visibility == null || visibility == post.visibility) return;
    try {
      await _repo.updatePost(post.id, visibility: visibility);
      _loadMyPosts();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('已更新可见范围')));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('更新失败: $e')));
      }
    }
  }

  String _visibilityLabel(String v) {
    switch (v) {
      case 'public': return '所有人';
      case 'followers': return '粉丝';
      case 'private': return '仅自己';
      default: return v;
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
                              : post.videoUrl != null
                                  ? const SizedBox(width: 56, height: 56, child: Icon(Icons.video_file, size: 32, color: Colors.blue))
                                  : null,
                          title: Row(
                            children: [
                              Expanded(
                                child: Text(post.title.isNotEmpty ? post.title : post.content, maxLines: 2, overflow: TextOverflow.ellipsis),
                              ),
                              const SizedBox(width: 8),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                decoration: BoxDecoration(
                                  color: Colors.grey.shade200,
                                  borderRadius: BorderRadius.circular(4),
                                ),
                                child: Text(
                                  _visibilityLabel(post.visibility),
                                  style: const TextStyle(fontSize: 10, color: Colors.grey),
                                ),
                              ),
                            ],
                          ),
                          subtitle: Text('${post.likeCount} 赞  ${post.commentCount} 评论  ${post.status}',
                              style: const TextStyle(fontSize: 12, color: Colors.grey)),
                          trailing: PopupMenuButton<String>(
                            onSelected: (action) {
                              switch (action) {
                                case 'edit': _editPost(post); break;
                                case 'visibility': _changeVisibility(post); break;
                                case 'delete': _deletePost(post.id); break;
                              }
                            },
                            itemBuilder: (_) => [
                              const PopupMenuItem(value: 'edit', child: Text('编辑')),
                              const PopupMenuItem(value: 'visibility', child: Text('修改可见范围')),
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
