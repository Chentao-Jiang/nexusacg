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
    final coverUrl = hasImage ? post.images.first : null;

    return GestureDetector(
      onTap: onTap,
      child: Card(
        clipBehavior: Clip.antiAlias,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        margin: EdgeInsets.zero,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            // Cover image / video thumbnail
            if (coverUrl != null)
              Stack(
                children: [
                  CachedNetworkImage(
                    imageUrl: coverUrl,
                    fit: BoxFit.cover,
                    width: double.infinity,
                    placeholder: (_, __) => Container(
                      color: Colors.grey.shade100,
                    ),
                    errorWidget: (_, __, ___) => Container(
                      height: 120,
                      color: Colors.grey.shade100,
                      child: const Icon(Icons.broken_image, color: Colors.grey),
                    ),
                  ),
                  // Play icon overlay for video posts
                  if (hasVideo)
                    Positioned(
                      left: 6, bottom: 6,
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                        decoration: BoxDecoration(
                          color: Colors.black54,
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: const Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(Icons.play_arrow, size: 14, color: Colors.white),
                            SizedBox(width: 2),
                            Text('视频', style: TextStyle(color: Colors.white, fontSize: 10)),
                          ],
                        ),
                      ),
                    ),
                  // Like count badge
                  if (post.likeCount > 0)
                    Positioned(
                      right: 4, top: 4,
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 2),
                        decoration: BoxDecoration(
                          color: Colors.black38,
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(Icons.favorite, size: 10, color: Colors.white),
                            const SizedBox(width: 3),
                            Text('${post.likeCount}', style: const TextStyle(color: Colors.white, fontSize: 10)),
                          ],
                        ),
                      ),
                    ),
                ],
              )
            else if (hasVideo)
              // Legacy: video without thumbnail
              Container(
                height: 120,
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [Color(0xFF2D2D3F), Color(0xFF1A1A2E)],
                  ),
                ),
                child: const Center(
                  child: Icon(Icons.play_circle_filled, size: 40, color: Colors.white54),
                ),
              )
            else
              // Text-only card
              Container(
                padding: const EdgeInsets.all(12),
                constraints: const BoxConstraints(minHeight: 70),
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [Color(0xFFFFF5E6), Color(0xFFFFF0D0)],
                  ),
                ),
                child: Text(
                  post.content,
                  maxLines: 3,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontSize: 13, height: 1.5, color: Color(0xFF5D4037)),
                ),
              ),
            // Footer
            Padding(
              padding: const EdgeInsets.fromLTRB(10, 8, 10, 10),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (post.title.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 4),
                      child: Text(
                        post.title,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13, height: 1.3),
                      ),
                    ),
                  Row(
                    children: [
                      CircleAvatar(
                        radius: 9,
                        backgroundImage: post.author?.avatarUrl != null
                            ? CachedNetworkImageProvider(post.author!.avatarUrl!)
                            : null,
                        child: post.author?.avatarUrl == null ? const Icon(Icons.person, size: 11) : null,
                      ),
                      const SizedBox(width: 5),
                      Expanded(
                        child: Text(
                          post.author?.nickname ?? '用户',
                          style: const TextStyle(color: Color(0xFF999999), fontSize: 11),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      const Icon(Icons.favorite_border, size: 13, color: Color(0xFFCCCCCC)),
                      const SizedBox(width: 2),
                      Text('${post.likeCount}', style: const TextStyle(color: Color(0xFFCCCCCC), fontSize: 11)),
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
}
