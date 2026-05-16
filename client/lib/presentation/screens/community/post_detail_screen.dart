import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:video_player/video_player.dart';
import 'package:nexusacg/presentation/screens/community/comments_screen.dart';

class PostDetailScreen extends StatefulWidget {
  final PostModel post;
  const PostDetailScreen({super.key, required this.post});

  @override
  State<PostDetailScreen> createState() => _PostDetailScreenState();
}

class _PostDetailScreenState extends State<PostDetailScreen> {
  VideoPlayerController? _videoController;
  bool _videoInitialized = false;
  bool _playing = false;
  final _repo = PostRepository();
  late bool _liked;

  @override
  void initState() {
    super.initState();
    _liked = false;
    _initVideo();
  }

  void _initVideo() {
    if (widget.post.videoUrl != null) {
      _videoController = VideoPlayerController.networkUrl(Uri.parse(widget.post.videoUrl!));
      _videoController!.initialize().then((_) {
        if (mounted) {
          setState(() => _videoInitialized = true);
        }
      }).catchError((_) {
        // Video failed to load
      });
    }
  }

  @override
  void dispose() {
    _videoController?.dispose();
    super.dispose();
  }

  Future<void> _toggleLike() async {
    try {
      if (_liked) {
        await _repo.unlikePost(widget.post.id);
      } else {
        await _repo.likePost(widget.post.id);
      }
      setState(() => _liked = !_liked);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('操作失败，请重试')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('帖子详情')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Author
            Row(
              children: [
                CircleAvatar(
                  radius: 20,
                  backgroundImage: widget.post.author?.avatarUrl != null
                      ? CachedNetworkImageProvider(widget.post.author!.avatarUrl!)
                      : null,
                  child: widget.post.author?.avatarUrl == null ? const Icon(Icons.person) : null,
                ),
                const SizedBox(width: 12),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      widget.post.author?.nickname ?? '用户',
                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                    ),
                    Text(
                      _timeAgo(widget.post.createdAt),
                      style: const TextStyle(color: Colors.grey, fontSize: 12),
                    ),
                  ],
                ),
              ],
            ),

            // Title
            if (widget.post.title.isNotEmpty) ...[
              const SizedBox(height: 16),
              Text(
                widget.post.title,
                style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
              ),
            ],

            // Content
            const SizedBox(height: 16),
            Text(
              widget.post.content,
              style: const TextStyle(fontSize: 15, height: 1.8),
            ),

            // Video
            if (widget.post.videoUrl != null) ...[
              const SizedBox(height: 16),
              _buildVideoPlayer(),
            ],

            // Images
            if (widget.post.images.isNotEmpty) ...[
              const SizedBox(height: 16),
              _buildImageGrid(),
            ],

            // Tags
            if (widget.post.tags.isNotEmpty) ...[
              const SizedBox(height: 12),
              Wrap(
                spacing: 6,
                children: widget.post.tags.map((t) => Chip(
                  label: Text('#$t', style: const TextStyle(fontSize: 12)),
                  padding: EdgeInsets.zero,
                  visualDensity: VisualDensity.compact,
                )).toList(),
              ),
            ],

            const SizedBox(height: 24),
            const Divider(),

            // Action bar
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 12),
              child: Row(
                children: [
                  _actionButton(
                    _liked ? Icons.favorite : Icons.favorite_border,
                    '${widget.post.likeCount + (_liked ? 1 : 0)}',
                    _liked ? Theme.of(context).primaryColor : Colors.grey,
                    _toggleLike,
                  ),
                  const SizedBox(width: 32),
                  _actionButton(
                    Icons.chat_bubble_outline,
                    '${widget.post.commentCount}',
                    Colors.grey,
                    () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => CommentsScreen(
                            postId: widget.post.id,
                            initialCount: widget.post.commentCount,
                          ),
                        ),
                      );
                    },
                  ),
                  const Spacer(),
                  _actionButton(
                    Icons.share,
                    '分享',
                    Colors.grey,
                    () {},
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVideoPlayer() {
    if (!_videoInitialized) {
      return Container(
        height: 200,
        width: double.infinity,
        decoration: BoxDecoration(
          color: Colors.black,
          borderRadius: BorderRadius.circular(8),
        ),
        child: const Center(child: CircularProgressIndicator(color: Colors.white)),
      );
    }
    return Stack(
      children: [
        AspectRatio(
          aspectRatio: _videoController!.value.aspectRatio,
          child: VideoPlayer(_videoController!),
        ),
        Positioned.fill(
          child: Center(
            child: GestureDetector(
              onTap: () {
                setState(() {
                  if (_playing) {
                    _videoController!.pause();
                  } else {
                    _videoController!.play();
                  }
                  _playing = !_playing;
                });
              },
              child: Container(
                decoration: BoxDecoration(
                  color: Colors.black26,
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  _playing ? Icons.pause_circle_filled : Icons.play_circle_filled,
                  size: 64,
                  color: Colors.white70,
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildImageGrid() {
    if (widget.post.images.length == 1) {
      return ClipRRect(
        borderRadius: BorderRadius.circular(8),
        child: CachedNetworkImage(
          imageUrl: widget.post.images[0],
          fit: BoxFit.cover,
          width: double.infinity,
        ),
      );
    }
    return SizedBox(
      height: 200,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        itemCount: widget.post.images.length,
        itemBuilder: (context, index) {
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: CachedNetworkImage(
                imageUrl: widget.post.images[index],
                width: 200,
                fit: BoxFit.cover,
                placeholder: (_, __) => Container(width: 200, color: Colors.grey.shade200),
                errorWidget: (_, __, ___) => Container(
                  width: 200,
                  color: Colors.grey.shade200,
                  child: const Icon(Icons.error),
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _actionButton(IconData icon, String label, Color color, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      child: Row(
        children: [
          Icon(icon, size: 22, color: color),
          const SizedBox(width: 4),
          Text(label, style: TextStyle(color: color, fontSize: 14)),
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
