import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/core/repositories/follow_repository.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/presentation/screens/community/comments_screen.dart';
import 'package:video_player/video_player.dart';
import 'dart:io';

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
  late bool _bookmarked;
  late bool _isFollowing;
  int _currentImageIndex = 0;
  final PageController _pageController = PageController();

  bool get _isOwnPost {
    // Check if current user is the author
    return false; // Simplified — will be overridden by the hidden follow button check
  }

  @override
  void initState() {
    super.initState();
    _liked = false;
    _collected = false;
    _bookmarked = false;
    _isFollowing = false;
    _initVideo();
    _checkFollow();
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

  @override
  void dispose() {
    _videoController?.dispose();
    _pageController.dispose();
    super.dispose();
  }

  Future<void> _toggleLike() async {
    try {
      if (_liked) { await _repo.unlikePost(widget.post.id); }
      else { await _repo.likePost(widget.post.id); }
      setState(() => _liked = !_liked);
    } catch (_) {}
  }

  Future<void> _toggleBookmark() async {
    try {
      if (_bookmarked) {
        await ApiClient().delete('/posts/${widget.post.id}/bookmark');
        if (mounted) setState(() => _bookmarked = false);
      } else {
        await ApiClient().post('/posts/${widget.post.id}/bookmark');
        if (mounted) setState(() => _bookmarked = true);
      }
    } catch (_) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('操作失败')));
    }
  }

  void _showShareSheet() {
    showModalBottomSheet(
      context: context,
      builder: (ctx) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 20),
          child: Wrap(
            children: [
              ListTile(
                leading: const Icon(Icons.copy, color: Colors.blue),
                title: const Text('复制链接'),
                onTap: () {
                  Navigator.pop(ctx);
                  Clipboard.setData(ClipboardData(text: 'http://101.133.169.72:8080/posts/${widget.post.id}'));
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('链接已复制')));
                },
              ),
              ListTile(
                leading: const Icon(Icons.wechat, color: Colors.green),
                title: const Text('分享到微信'),
                onTap: () {
                  Navigator.pop(ctx);
                  Clipboard.setData(ClipboardData(text: 'http://101.133.169.72:8080/posts/${widget.post.id}'));
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('链接已复制，请打开微信粘贴分享')));
                },
              ),
              if (widget.post.images.isNotEmpty)
                ListTile(
                  leading: const Icon(Icons.save_alt, color: Colors.orange),
                  title: const Text('保存图片'),
                  onTap: () {
                    Navigator.pop(ctx);
                    _saveCurrentImage();
                  },
                ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _saveCurrentImage() async {
    if (widget.post.images.isEmpty) return;
    final url = widget.post.images[_currentImageIndex];
    try {
      final dir = Directory.systemTemp;
      final file = File('${dir.path}/nexusacg_save_${DateTime.now().millisecondsSinceEpoch}.jpg');
      // Use http to download
      final httpClient = HttpClient();
      final request = await httpClient.getUrl(Uri.parse(url));
      final response = await request.close();
      await response.pipe(file.openWrite());
      httpClient.close();
      await file.writeAsBytes(await file.readAsBytes());
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('图片已保存到临时目录')));
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('保存失败: $e')));
    }
  }

  void _showImageFullscreen(int index) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => _ImageViewer(
          images: widget.post.images,
          initialIndex: index,
        ),
      ),
    );
  }

  void _showMoreMenu() {
    showModalBottomSheet(
      context: context,
      builder: (ctx) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 20),
          child: Wrap(
            children: [
              ListTile(
                leading: const Icon(Icons.share_outlined),
                title: const Text('分享'),
                onTap: () { Navigator.pop(ctx); _showShareSheet(); },
              ),
              ListTile(
                leading: const Icon(Icons.report_outlined, color: Colors.red),
                title: const Text('举报'),
                onTap: () {
                  Navigator.pop(ctx);
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('举报已提交')));
                },
              ),
              ListTile(
                leading: const Icon(Icons.block_outlined),
                title: const Text('不感兴趣'),
                onTap: () {
                  Navigator.pop(ctx);
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('将减少此类推荐')));
                },
              ),
            ],
          ),
        ),
      ),
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
                child: widget.post.author?.avatarUrl == null ? const Icon(Icons.person, size: 16) : null,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(widget.post.author?.nickname ?? '用户',
                      style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Colors.black87)),
                  Text(_timeAgo(widget.post.createdAt),
                      style: const TextStyle(fontSize: 11, color: Colors.grey)),
                ],
              ),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.more_horiz, color: Colors.black87),
            onPressed: _showMoreMenu,
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
                  // Media area — video OR swipeable images
                  if (hasVideo) _buildMediaPlayer(),
                  if (hasImage && !hasVideo) _buildImageGallery(),

                  // Content area
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (widget.post.title.isNotEmpty) ...[
                          Text(widget.post.title,
                              style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, height: 1.3)),
                          const SizedBox(height: 8),
                        ],
                        if (widget.post.content.isNotEmpty) ...[
                          Text(widget.post.content,
                              style: const TextStyle(fontSize: 15, height: 1.7, color: Colors.black87)),
                          const SizedBox(height: 10),
                        ],
                        if (widget.post.tags.isNotEmpty) ...[
                          Wrap(
                            spacing: 8, runSpacing: 4,
                            children: widget.post.tags.map((t) => Text('#$t',
                                style: const TextStyle(fontSize: 13, color: Color(0xFF5974A8)))).toList(),
                          ),
                          const SizedBox(height: 10),
                        ],
                        Text(_timeAgo(widget.post.createdAt),
                            style: const TextStyle(fontSize: 12, color: Colors.grey)),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                  const Divider(height: 1),

                  // Like count bar
                  if (widget.post.likeCount > 0) ...[
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                      child: Row(children: [
                        const Icon(Icons.favorite, size: 18, color: Colors.red),
                        const SizedBox(width: 6),
                        Text('${widget.post.likeCount + (_liked ? 1 : 0)} 人赞了',
                            style: const TextStyle(fontSize: 13, color: Colors.black54)),
                      ]),
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
                    onTap: () => _openComments(),
                  ),
                  _bottomAction(
                    icon: _bookmarked ? Icons.bookmark : Icons.bookmark_border,
                    label: _bookmarked ? '已收藏' : '收藏',
                    color: _bookmarked ? Colors.amber : Colors.black54,
                    onTap: _toggleBookmark,
                  ),
                  _bottomAction(
                    icon: Icons.share_outlined,
                    label: '分享',
                    color: Colors.black54,
                    onTap: _showShareSheet,
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ── Video player with BoxFit.contain ──
  Widget _buildMediaPlayer() {
    if (_videoError) {
      return Container(
        height: MediaQuery.of(context).size.width * 0.56,
        width: double.infinity,
        color: Colors.black87,
        child: Center(
          child: Column(mainAxisSize: MainAxisSize.min, children: [
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
              label: const Text('重试'),
            ),
          ]),
        ),
      );
    }
    if (!_videoInitialized) {
      return Container(
        height: MediaQuery.of(context).size.width * 0.56,
        width: double.infinity,
        color: Colors.black,
        child: const Center(child: CircularProgressIndicator(color: Colors.white)),
      );
    }

    return GestureDetector(
      onTap: () {
        setState(() {
          _playing ? _videoController!.pause() : _videoController!.play();
          _playing = !_playing;
        });
      },
      child: Container(
        color: Colors.black,
        width: double.infinity,
        child: AspectRatio(
          aspectRatio: _videoController!.value.aspectRatio,
          child: Stack(
            alignment: Alignment.center,
            children: [
              FittedBox(
                fit: BoxFit.contain,
                child: SizedBox(
                  width: _videoController!.value.size.width,
                  height: _videoController!.value.size.height,
                  child: VideoPlayer(_videoController!),
                ),
              ),
              if (!_playing)
                Container(
                  width: 56, height: 56,
                  decoration: const BoxDecoration(color: Colors.black38, shape: BoxShape.circle),
                  child: const Icon(Icons.play_arrow, size: 36, color: Colors.white),
                ),
            ],
          ),
        ),
      ),
    );
  }

  // ── Swipeable image gallery ──
  Widget _buildImageGallery() {
    final images = widget.post.images;
    final screenWidth = MediaQuery.of(context).size.width;
    return Column(
      children: [
        SizedBox(
          height: screenWidth,
          width: double.infinity,
          child: PageView.builder(
            controller: _pageController,
            onPageChanged: (i) => setState(() => _currentImageIndex = i),
            itemCount: images.length,
            itemBuilder: (ctx, i) => GestureDetector(
              onTap: () => _showImageFullscreen(i),
              onLongPress: () => _saveImageToGallery(images[i]),
              child: Container(
                color: Colors.white,
                child: Center(
                  child: CachedNetworkImage(
                    imageUrl: images[i],
                    fit: BoxFit.contain,
                    width: screenWidth,
                    placeholder: (_, __) => Container(color: Colors.grey.shade100, child: const Center(child: CircularProgressIndicator())),
                    errorWidget: (_, __, ___) => Container(color: Colors.grey.shade100, child: const Icon(Icons.broken_image, color: Colors.grey, size: 48)),
                  ),
                ),
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

  Future<void> _saveImageToGallery(String url) async {
    try {
      final dir = Directory.systemTemp;
      final file = File('${dir.path}/nexusacg_img_${DateTime.now().millisecondsSinceEpoch}.jpg');
      final httpClient = HttpClient();
      final request = await httpClient.getUrl(Uri.parse(url));
      final response = await request.close();
      await response.pipe(file.openWrite());
      httpClient.close();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('图片已保存')));
      }
    } catch (_) {}
  }

  // ── Comments ──
  Widget _buildCommentsSection() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('评论', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          if (widget.post.commentCount == 0)
            const Center(child: Padding(
              padding: EdgeInsets.all(24),
              child: Text('暂无评论，来说点什么吧', style: TextStyle(color: Colors.grey, fontSize: 13)),
            )),
          if (widget.post.commentCount > 0)
            GestureDetector(
              onTap: _openComments,
              child: Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(color: Colors.grey.shade50, borderRadius: BorderRadius.circular(8)),
                child: Row(children: [
                  const Icon(Icons.chat_bubble_outline, size: 18, color: Colors.grey),
                  const SizedBox(width: 8),
                  Text('查看全部 ${widget.post.commentCount} 条评论', style: const TextStyle(fontSize: 13, color: Colors.grey)),
                  const Spacer(),
                  const Icon(Icons.chevron_right, size: 18, color: Colors.grey),
                ]),
              ),
            ),
        ],
      ),
    );
  }

  void _openComments() {
    Navigator.push(context, MaterialPageRoute(
      builder: (_) => CommentsScreen(postId: widget.post.id, initialCount: widget.post.commentCount),
    ));
  }

  Widget _bottomAction({required IconData icon, required String label, required Color color, VoidCallback? onTap}) {
    return InkWell(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          Icon(icon, size: 24, color: color),
          const SizedBox(height: 2),
          Text(label, style: TextStyle(fontSize: 11, color: color)),
        ]),
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

// ── Fullscreen image viewer with pinch-to-zoom ──
class _ImageViewer extends StatelessWidget {
  final List<String> images;
  final int initialIndex;
  const _ImageViewer({required this.images, required this.initialIndex});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
      ),
      body: PageView.builder(
        controller: PageController(initialPage: initialIndex),
        itemCount: images.length,
        itemBuilder: (_, i) => InteractiveViewer(
          minScale: 0.5,
          maxScale: 4.0,
          child: Center(
            child: CachedNetworkImage(
              imageUrl: images[i],
              fit: BoxFit.contain,
              placeholder: (_, __) => const Center(child: CircularProgressIndicator(color: Colors.white)),
              errorWidget: (_, __, ___) => const Icon(Icons.broken_image, color: Colors.white54, size: 64),
            ),
          ),
        ),
      ),
    );
  }
}
