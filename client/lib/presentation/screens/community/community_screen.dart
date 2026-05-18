import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter_staggered_grid_view/flutter_staggered_grid_view.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/screens/community/post_detail_screen.dart';
import 'package:nexusacg/presentation/screens/community/post_create_screen.dart';
import 'package:nexusacg/presentation/screens/community/my_posts_screen.dart';

class CommunityScreen extends StatefulWidget {
  const CommunityScreen({super.key});

  @override
  State<CommunityScreen> createState() => _CommunityScreenState();
}

class _CommunityScreenState extends State<CommunityScreen> {
  final _repo = PostRepository();
  List<PostModel> _posts = [];
  bool _loading = true;
  int _page = 1;
  bool _hasMore = true;

  @override
  void initState() {
    super.initState();
    _loadPosts();
  }

  Future<void> _loadPosts() async {
    setState(() => _loading = true);
    _page = 1;
    try {
      final posts = await _repo.getPosts(page: _page);
      setState(() {
        _posts = posts;
        _hasMore = posts.length >= 20;
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加载失败: $e')),
        );
      }
    }
  }

  Future<void> _loadMore() async {
    if (!_hasMore) return;
    final nextPage = _page + 1;
    try {
      final posts = await _repo.getPosts(page: nextPage);
      setState(() {
        _posts.addAll(posts);
        _page = nextPage;
        _hasMore = posts.length >= 20;
      });
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('社区'),
        actions: [
          IconButton(
            icon: const Icon(Icons.article_outlined),
            tooltip: '我的帖子',
            onPressed: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const MyPostsScreen())),
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _loadPosts,
              child: _posts.isEmpty
                  ? const Center(child: Text('暂无内容，快来发第一条帖子吧！'))
                  : NotificationListener<ScrollNotification>(
                      onNotification: (notification) {
                        if (notification is ScrollEndNotification &&
                            notification.metrics.pixels >= notification.metrics.maxScrollExtent * 0.8) {
                          _loadMore();
                        }
                        return false;
                      },
                      child: MasonryGridView.count(
                        crossAxisCount: 2,
                        mainAxisSpacing: 8,
                        crossAxisSpacing: 8,
                        padding: const EdgeInsets.all(8),
                        itemCount: _posts.length,
                        itemBuilder: (context, index) => _XhsCard(_posts[index], () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(builder: (_) => PostDetailScreen(post: _posts[index])),
                          );
                        }),
                      ),
                    ),
            ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const PostCreateScreen()),
          ).then((_) => _loadPosts());
        },
        child: const Icon(Icons.add),
      ),
    );
  }
}

class _XhsCard extends StatelessWidget {
  final PostModel post;
  final VoidCallback onTap;

  const _XhsCard(this.post, this.onTap);

  @override
  Widget build(BuildContext context) {
    final hasImage = post.images.isNotEmpty;
    final hasVideo = post.videoUrl != null;
    return GestureDetector(
      onTap: onTap,
      child: Card(
        clipBehavior: Clip.antiAlias,
        elevation: 1,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Cover image or text card
            if (hasImage)
              CachedNetworkImage(
                imageUrl: post.images.first,
                fit: BoxFit.cover,
                width: double.infinity,
                placeholder: (_, __) => Container(
                  height: 150 + (indexOf(0) % 3) * 40.0,
                  color: Colors.grey.shade200,
                ),
                errorWidget: (_, __, ___) => Container(
                  height: 150,
                  color: Colors.grey.shade200,
                  child: const Icon(Icons.broken_image, color: Colors.grey),
                ),
              )
            else if (hasVideo)
              Container(
                height: 150,
                color: Colors.black12,
                child: const Center(
                  child: Icon(Icons.play_circle_fill, size: 48, color: Colors.white70),
                ),
              )
            else
              Container(
                padding: const EdgeInsets.all(12),
                child: Text(
                  post.content,
                  maxLines: 3,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontSize: 14, height: 1.4),
                ),
              ),
            // Bottom info
            Padding(
              padding: const EdgeInsets.all(8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (post.title.isNotEmpty)
                    Text(post.title, maxLines: 1, overflow: TextOverflow.ellipsis,
                        style: const TextStyle(fontWeight: FontWeight.w500, fontSize: 13)),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      CircleAvatar(
                        radius: 10,
                        backgroundImage: post.author?.avatarUrl != null
                            ? CachedNetworkImageProvider(post.author!.avatarUrl!)
                            : null,
                        child: post.author?.avatarUrl == null ? const Icon(Icons.person, size: 12) : null,
                      ),
                      const SizedBox(width: 4),
                      Expanded(
                        child: Text(post.author?.nickname ?? '用户',
                            style: const TextStyle(color: Colors.grey, fontSize: 11),
                            overflow: TextOverflow.ellipsis),
                      ),
                      const Icon(Icons.favorite_border, size: 12, color: Colors.grey),
                      const SizedBox(width: 2),
                      Text('${post.likeCount}', style: const TextStyle(color: Colors.grey, fontSize: 11)),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  int indexOf(int fallback) => 0;
}
