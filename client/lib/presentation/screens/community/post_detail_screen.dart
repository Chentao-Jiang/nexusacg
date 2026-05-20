import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/core/repositories/follow_repository.dart';
import 'package:video_player/video_player.dart';

class PostDetailScreen extends StatefulWidget {
  final PostModel post;
  const PostDetailScreen({super.key, required this.post});

  @override
  State<PostDetailScreen> createState() => _PostDetailScreenState();
}

class _PostDetailScreenState extends State<PostDetailScreen> {
  VideoPlayerController? _videoController;
  bool _videoInitialized = false;
  bool _videoError = false;
  String _videoErrorMessage = '';
  bool _playing = false;
  final _repo = PostRepository();
  final _followRepo = FollowRepository();
  late bool _liked;
  late bool _collected;
  late bool _isFollowing;
  int _currentImageIndex = 0;

  @override
  void initState() {
    super.initState();
    _liked = false;
    _collected = false;
    _isFollowing = false;
    _initVideo();
    _checkFollow();
  }

  Future<void> _checkFollow() async {
    if (widget.post.author?.id != null) {
      final following = await _followRepo.isFollowing(widget.post.author!.id!);
      if (mounted) setState(() => _isFollowing = following);
    }
  }

  Future<void> _toggleFollow() async {
    final authorId = widget.post.author?.id;
    if (authorId == null) return;
    if (_isFollowing) {
      await _followRepo.unfollow(authorId);
      if (mounted) setState(() => _isFollowing = false);
    } else {
      await _followRepo.follow(authorId);
      if (mounted) setState(() => _isFollowing = true);
    }
  }

  void _initVideo() {
    if (widget.post.videoUrl != null) {
      _videoController = VideoPlayerController.networkUrl(Uri.parse(widget.post.videoUrl!));
      _videoController!.initialize().then((_) {
        if (mounted) {
          setState(() => _videoInitialized = true);
          _videoController!.setLooping(true);
        }
      }).catchError((error) {
        if (mounted) {
          setState(() {
            _videoError = true;
            _videoErrorMessage = error.toString();
          });
        }
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
    } catch (_) {}
  }

  void _toggleCollect() {
    setState(() => _collected = !_collected);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(_collected ? '已收藏' : '已取消收藏')),
    );
  }

  @override
  Widget build(BuildContext context) {
    final hasImage = widget.post.images.isNotEmpty;
    final hasVideo = widget.post.videoUrl != null;

    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        leading: const BackButton(color: Colors.black87),
        titleSpacing: 0,
        title: Row(
          children: [
            GestureDetector(
              child: CircleAvatar(
                radius: 13,
                backgroundImage: widget.post.author?.avatarUrl != null
                    ? CachedNetworkImageProvider(widget.post.author!.avatarUrl!)
                    : null,
                child: widget.post.author?.avatarUrl == null
                    ? const Icon(Icons.person, size: 16)
                    : null,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    widget.post.author?.nickname ?? '用户',
                    style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Colors.black87),
                  ),
                  Text(
                    _timeAgo(widget.post.createdAt),
                    style: const TextStyle(fontSize: 11, color: Colors.grey),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            OutlinedButton(
              onPressed: _toggleFollow,
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                minimumSize: Size.zero,
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                side: BorderSide(color: _isFollowing ? Colors.grey : Colors.red),
                foregroundColor: _isFollowing ? Colors.grey : Colors.red,
                textStyle: const TextStyle(fontSize: 12),
              ),
              child: Text(_isFollowing ? '已关注' : '关注'),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.more_horiz, color: Colors.black87),
            onPressed: () {},
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Media area
                  if (hasVideo) _buildMediaPlayer(),
                  if (hasImage) _buildImageGallery(),

                  // Content area
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (widget.post.title.isNotEmpty) ...[
                          Text(
                            widget.post.title,
                            style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, height: 1.3),
                          ),
                          const SizedBox(height: 8),
                        ],
                        if (widget.post.content.isNotEmpty) ...[
                          Text(
                            widget.post.content,
                            style: const TextStyle(fontSize: 15, height: 1.7, color: Colors.black87),
                          ),
                          const SizedBox(height: 10),
                        ],
                        if (widget.post.tags.isNotEmpty) ...[
                          Wrap(
                            spacing: 8,
                            runSpacing: 4,
                            children: widget.post.tags.map((t) => Text(
                              '#$t',
                              style: const TextStyle(fontSize: 13, color: Color(0xFF5974A8)),
                            )).toList(),
                          ),
                          const SizedBox(height: 10),
                        ],
                        Text(
                          _timeAgo(widget.post.createdAt),
                          style: const TextStyle(fontSize: 12, color: Colors.grey),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 16),
                  const Divider(height: 1),

                  // Like count bar
                  if (widget.post.likeCount > 0) ...[
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                      child: Row(
                        children: [
                          const Icon(Icons.favorite, size: 18, color: Colors.red),
                          const SizedBox(width: 6),
                          Text(
                            '${widget.post.likeCount + (_liked ? 1 : 0)} 人赞了',
                            style: const TextStyle(fontSize: 13, color: Colors.black54),
                          ),
                        ],
                      ),
                    ),
                    const Divider(height: 1),
                  ],

                  // Comments section
                  _buildCommentsSection(),
                ],
              ),
            ),
          ),

          // Fixed bottom bar
          Container(
            decoration: BoxDecoration(
              color: Colors.white,
              border: Border(top: BorderSide(color: Colors.grey.shade200)),
            ),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            child: SafeArea(
              top: false,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: [
                  _bottomAction(
                    icon: _liked ? Icons.favorite : Icons.favorite_border,
                    label: '${widget.post.likeCount + (_liked ? 1 : 0)}',
                    color: _liked ? Colors.red : Colors.black54,
                    onTap: _toggleLike,
                  ),
                  _bottomAction(
                    icon: Icons.chat_bubble_outline,
                    label: '${widget.post.commentCount}',
                    color: Colors.black54,
                    onTap: () => _scrollToComments(),
                  ),
                  _bottomAction(
                    icon: _collected ? Icons.star : Icons.star_border,
                    label: _collected ? '已收藏' : '收藏',
                    color: _collected ? Colors.amber : Colors.black54,
                    onTap: _toggleCollect,
                  ),
                  _bottomAction(
                    icon: Icons.share_outlined,
                    label: '分享',
                    color: Colors.black54,
                    onTap: () {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('分享功能开发中')),
                      );
                    },
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMediaPlayer() {
    if (_videoError) {
      return Container(
        height: MediaQuery.of(context).size.width * 0.75,
        width: double.infinity,
        color: Colors.black87,
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.error_outline, size: 48, color: Colors.white54),
              const SizedBox(height: 8),
              const Text('视频无法播放', style: TextStyle(color: Colors.white70)),
              const SizedBox(height: 12),
              ElevatedButton.icon(
                onPressed: () {
                  setState(() { _videoError = false; _videoInitialized = false; });
                  _initVideo();
                },
                icon: const Icon(Icons.refresh, size: 16),
                label: const Text('重试', style: TextStyle(fontSize: 12)),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
                ),
              ),
            ],
          ),
        ),
      );
    }
    if (!_videoInitialized) {
      return Container(
        height: MediaQuery.of(context).size.width * 0.75,
        width: double.infinity,
        color: Colors.black,
        child: const Center(child: CircularProgressIndicator(color: Colors.white)),
      );
    }

    final videoAspect = _videoController!.value.aspectRatio;
    final screenWidth = MediaQuery.of(context).size.width;

    return GestureDetector(
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
      child: Stack(
        alignment: Alignment.center,
        children: [
          ClipRRect(
            child: SizedBox(
              width: screenWidth,
              height: screenWidth / videoAspect,
              child: FittedBox(
                fit: BoxFit.cover,
                child: SizedBox(
                  width: _videoController!.value.size.width,
                  height: _videoController!.value.size.height,
                  child: VideoPlayer(_videoController!),
                ),
              ),
            ),
          ),
          if (!_playing)
            Container(
              width: 56, height: 56,
              decoration: const BoxDecoration(
                color: Colors.black38, shape: BoxShape.circle,
              ),
              child: const Icon(Icons.play_arrow, size: 36, color: Colors.white),
            ),
        ],
      ),
    );
  }

  Widget _buildImageGallery() {
    final images = widget.post.images;
    return Column(
      children: [
        SizedBox(
          height: MediaQuery.of(context).size.width * 0.85,
          width: double.infinity,
          child: PageView.builder(
            onPageChanged: (i) => setState(() => _currentImageIndex = i),
            itemCount: images.length,
            itemBuilder: (ctx, i) => CachedNetworkImage(
              imageUrl: images[i],
              fit: BoxFit.cover,
              width: double.infinity,
              placeholder: (_, __) => Container(color: Colors.grey.shade100),
              errorWidget: (_, __, ___) => Container(
                color: Colors.grey.shade100,
                child: const Icon(Icons.broken_image, color: Colors.grey, size: 48),
              ),
            ),
          ),
        ),
        if (images.length > 1)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: List.generate(images.length, (i) => Container(
                width: i == _currentImageIndex ? 16 : 6,
                height: 6,
                margin: const EdgeInsets.symmetric(horizontal: 3),
                decoration: BoxDecoration(
                  color: i == _currentImageIndex ? Colors.red : Colors.grey.shade300,
                  borderRadius: BorderRadius.circular(3),
                ),
              )),
            ),
          ),
      ],
    );
  }

  Widget _buildCommentsSection() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('评论', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          if (widget.post.commentCount == 0)
            const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Text('暂无评论，来说点什么吧', style: TextStyle(color: Colors.grey, fontSize: 13)),
              ),
            ),
          // Show first 3 comments inline
          if (widget.post.commentCount > 0)
            GestureDetector(
              onTap: () => _openComments(),
              child: Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.grey.shade50,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.chat_bubble_outline, size: 18, color: Colors.grey),
                    const SizedBox(width: 8),
                    Text(
                      '查看全部 ${widget.post.commentCount} 条评论',
                      style: const TextStyle(fontSize: 13, color: Colors.grey),
                    ),
                    const Spacer(),
                    const Icon(Icons.chevron_right, size: 18, color: Colors.grey),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }

  void _scrollToComments() {
    // Navigate to comments screen for now
    _openComments();
  }

  void _openComments() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => CommentsScreen(
          postId: widget.post.id,
          initialCount: widget.post.commentCount,
        ),
      ),
    );
  }

  Widget _bottomAction({required IconData icon, required String label, required Color color, VoidCallback? onTap}) {
    return InkWell(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 24, color: color),
            const SizedBox(height: 2),
            Text(label, style: TextStyle(fontSize: 11, color: color)),
          ],
        ),
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
