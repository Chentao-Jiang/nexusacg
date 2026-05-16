import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/screens/community/post_detail_screen.dart';
import 'package:nexusacg/presentation/screens/community/post_create_screen.dart';

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
    } catch (e) {
      // Ignore load more errors
    }
  }

  void _goToDetail(PostModel post) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => PostDetailScreen(post: post)),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('社区')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _loadPosts,
              child: _posts.isEmpty
                  ? const Center(child: Text('暂无内容，快来发第一条帖子吧！'))
                  : NotificationListener<ScrollNotification>(
                      onNotification: (notification) {
                        if (notification is ScrollEndNotification &&
                            notification.metrics.pixels >=
                                notification.metrics.maxScrollExtent * 0.8) {
                          _loadMore();
                        }
                        return false;
                      },
                      child: ListView.builder(
                        padding: const EdgeInsets.all(12),
                        itemCount: _posts.length,
                        itemBuilder: (context, index) => GestureDetector(
                          onTap: () => _goToDetail(_posts[index]),
                          child: _PostCard(_posts[index]),
                        ),
                      ),
                    ),
            ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const PostCreateScreen()),
          ).then((_) => _loadPosts()); // Refresh after post
        },
        child: const Icon(Icons.edit),
      ),
    );
  }
}

class _PostCard extends StatelessWidget {
  final PostModel post;
  const _PostCard(this.post);

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                CircleAvatar(
                  radius: 16,
                  backgroundImage: post.author?.avatarUrl != null
                      ? CachedNetworkImageProvider(post.author!.avatarUrl!)
                      : null,
                  child: post.author?.avatarUrl == null ? const Icon(Icons.person) : null,
                ),
                const SizedBox(width: 8),
                Text(
                  post.author?.nickname ?? '用户',
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
                const Spacer(),
                Text(_timeAgo(post.createdAt), style: const TextStyle(color: Colors.grey, fontSize: 12)),
              ],
            ),
            if (post.title.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(post.title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            ],
            const SizedBox(height: 8),
            Text(post.content, maxLines: 5, overflow: TextOverflow.ellipsis),

            // Video thumbnail
            if (post.videoUrl != null) ...[
              const SizedBox(height: 8),
              Stack(
                children: [
                  Container(
                    height: 200,
                    width: double.infinity,
                    decoration: BoxDecoration(
                      color: Colors.black,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Center(
                      child: Icon(Icons.play_circle_outline, size: 48, color: Colors.white70),
                    ),
                  ),
                  Positioned(
                    bottom: 8,
                    right: 8,
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: Colors.black54,
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: const Text(
                        '视频',
                        style: TextStyle(color: Colors.white, fontSize: 12),
                      ),
                    ),
                  ),
                ],
              ),
            ],

            // Images
            if (post.images.isNotEmpty) ...[
              const SizedBox(height: 8),
              SizedBox(
                height: 180,
                child: ListView.builder(
                  scrollDirection: Axis.horizontal,
                  itemCount: post.images.length,
                  itemBuilder: (context, index) {
                    return Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: CachedNetworkImage(
                          imageUrl: post.images[index],
                          width: 180,
                          fit: BoxFit.cover,
                          placeholder: (_, __) => Container(width: 180, color: Colors.grey.shade200),
                          errorWidget: (_, __, ___) => Container(width: 180, color: Colors.grey.shade200, child: const Icon(Icons.error)),
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
            if (post.tags.isNotEmpty) ...[
              const SizedBox(height: 8),
              Wrap(
                spacing: 6,
                children: post.tags.map((t) => Chip(
                  label: Text('#$t', style: const TextStyle(fontSize: 11)),
                  padding: EdgeInsets.zero,
                  visualDensity: VisualDensity.compact,
                )).toList(),
              ),
            ],
            const SizedBox(height: 12),
            Row(
              children: [
                _actionButton(Icons.favorite_border, '${post.likeCount}'),
                const SizedBox(width: 24),
                _actionButton(Icons.chat_bubble_outline, '${post.commentCount}'),
                const Spacer(),
                _actionButton(Icons.share, '分享'),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _actionButton(IconData icon, String label) {
    return InkWell(
      onTap: () {},
      child: Row(
        children: [
          Icon(icon, size: 18, color: Colors.grey),
          const SizedBox(width: 4),
          Text(label, style: const TextStyle(color: Colors.grey, fontSize: 13)),
        ],
      ),
    );
  }

  String _timeAgo(DateTime dt) {
    final diff = DateTime.now().difference(dt);
    if (diff.inMinutes < 1) return '刚刚';
    if (diff.inMinutes < 60) return '${diff.inMinutes}分钟前';
    if (diff.inHours < 24) return '${diff.inHours}小时前';
    return '${diff.inDays}天前';
  }
}
