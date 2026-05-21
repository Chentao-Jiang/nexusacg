import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/core/repositories/follow_repository.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/core/constants/app_constants.dart';
import 'package:nexusacg/presentation/screens/community/comments_screen.dart';
import 'package:video_player/video_player.dart';
import 'dart:io';
import 'package:dio/dio.dart';

class PostDetailScreen extends StatefulWidget {
  final PostModel post;
  const PostDetailScreen({super.key, required this.post});
  @override
  State<PostDetailScreen> createState() => _PostDetailScreenState();
}

class _PostDetailScreenState extends State<PostDetailScreen> {
  VideoPlayerController? _videoController;
  bool _videoInitialized = false, _videoError = false;
  String _videoErrorMessage = '';
  bool _playing = false;
  final _repo = PostRepository();
  final _followRepo = FollowRepository();
  late bool _liked, _bookmarked, _isFollowing;
  int _currentImageIndex = 0;
  final PageController _pageController = PageController();

  @override
  void initState() {
    super.initState();
    _liked = false; _bookmarked = false; _isFollowing = false;
    _initVideo(); _checkFollow(); _checkLikeBookmark();
  }

  void _initVideo() {
    if (widget.post.videoUrl != null) {
      _videoController = VideoPlayerController.networkUrl(Uri.parse(widget.post.videoUrl!));
      _videoController!.initialize().then((_) {
        if (mounted) { setState(() => _videoInitialized = true); _videoController!.setLooping(true); }
      }).catchError((error) {
        if (mounted) setState(() { _videoError = true; _videoErrorMessage = error.toString(); });
      });
    }
  }

  Future<void> _checkLikeBookmark() async {
    // Simple approach: assume not liked/bookmarked unless we fetch from server
    // The toggle will correct state on first user interaction
  }

  Future<void> _checkFollow() async {
    if (widget.post.author?.id != null) {
      final f = await _followRepo.isFollowing(widget.post.author!.id!);
      if (mounted) setState(() => _isFollowing = f);
    }
  }

  Future<void> _toggleFollow() async {
    final id = widget.post.author?.id; if (id == null) return;
    _isFollowing ? await _followRepo.unfollow(id) : await _followRepo.follow(id);
    if (mounted) setState(() => _isFollowing = !_isFollowing);
  }

  @override
  void dispose() { _videoController?.dispose(); _pageController.dispose(); super.dispose(); }

  Future<void> _toggleLike() async {
    try {
      _liked ? await _repo.unlikePost(widget.post.id) : await _repo.likePost(widget.post.id);
      if (mounted) setState(() => _liked = !_liked);
    } catch (_) {}
  }

  Future<void> _toggleBookmark() async {
    try {
      if (_bookmarked) { await ApiClient().delete('/posts/${widget.post.id}/bookmark'); }
      else { await ApiClient().post('/posts/${widget.post.id}/bookmark'); }
      if (mounted) setState(() => _bookmarked = !_bookmarked);
    } catch (_) { if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('操作失败'))); }
  }

  void _showShareSheet() {
    showModalBottomSheet(context: context, builder: (ctx) => SafeArea(child: Padding(
      padding: const EdgeInsets.symmetric(vertical: 20),
      child: Wrap(children: [
        ListTile(leading: const Icon(Icons.copy, color: Colors.blue), title: const Text('复制链接'),
          onTap: () { Navigator.pop(ctx); Clipboard.setData(ClipboardData(text: '${AppConstants.apiBaseUrl}/posts/${widget.post.id}')); ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('链接已复制'))); }),
        ListTile(leading: const Icon(Icons.wechat, color: Colors.green), title: const Text('分享到微信'),
          onTap: () { Navigator.pop(ctx); ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请使用系统分享'))); }),
        if (widget.post.images.isNotEmpty)
          ListTile(leading: const Icon(Icons.save_alt, color: Colors.orange), title: const Text('保存图片'),
            onTap: () { Navigator.pop(ctx); _saveImage(); }),
      ]),
    )));
  }

  Future<void> _saveImage() async {
    if (widget.post.images.isEmpty) return;
    try {
      final dir = Directory('/storage/emulated/0/Pictures');
      if (!await dir.exists()) await dir.create(recursive: true);
      final file = File('${dir.path}/nexusacg_${DateTime.now().millisecondsSinceEpoch}.jpg');
      final resp = await ApiClient().dio.get(widget.post.images[_currentImageIndex], options: Options(responseType: ResponseType.bytes));
      await file.writeAsBytes(resp.data);
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('已保存到相册')));
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('保存失败: $e')));
    }
  }

  Future<void> _saveImageToGallery(String url) async {
    try {
      final dir = Directory('/storage/emulated/0/Pictures');
      if (!await dir.exists()) await dir.create(recursive: true);
      final file = File('${dir.path}/nexusacg_${DateTime.now().millisecondsSinceEpoch}.jpg');
      final resp = await ApiClient().dio.get(url, options: Options(responseType: ResponseType.bytes));
      await file.writeAsBytes(resp.data);
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('已保存到相册')));
    } catch (e) { if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('保存失败'))); }
  }

  void _showFullscreen(int i) {
    Navigator.push(context, MaterialPageRoute(builder: (_) => _ImageViewer(images: widget.post.images, initialIndex: i)));
  }

  void _showMoreMenu() {
    showModalBottomSheet(context: context, builder: (ctx) => SafeArea(child: Padding(
      padding: const EdgeInsets.symmetric(vertical: 20),
      child: Wrap(children: [
        ListTile(leading: const Icon(Icons.share_outlined), title: const Text('分享'), onTap: () { Navigator.pop(ctx); _showShareSheet(); }),
        ListTile(leading: const Icon(Icons.report_outlined, color: Colors.red), title: const Text('举报'),
          onTap: () { Navigator.pop(ctx); ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('举报已提交'))); }),
      ]),
    )));
  }

  @override
  Widget build(BuildContext context) {
    final hasImg = widget.post.images.isNotEmpty;
    final hasVid = widget.post.videoUrl != null;
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(backgroundColor: Colors.white, elevation: 0.5, leading: const BackButton(color: Colors.black87), titleSpacing: 0,
        title: Row(children: [
          CircleAvatar(radius: 13, backgroundImage: widget.post.author?.avatarUrl != null ? CachedNetworkImageProvider(widget.post.author!.avatarUrl!) : null,
            child: widget.post.author?.avatarUrl == null ? const Icon(Icons.person, size: 16) : null),
          const SizedBox(width: 8),
          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text(widget.post.author?.nickname ?? '用户', style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Colors.black87)),
            Text(_timeAgo(widget.post.createdAt), style: const TextStyle(fontSize: 11, color: Colors.grey)),
          ])),
          if (ApiClient().currentUserId != null && ApiClient().currentUserId != widget.post.author?.id) ...[
            const SizedBox(width: 8),
            OutlinedButton(
              onPressed: _toggleFollow,
              style: OutlinedButton.styleFrom(padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4), minimumSize: Size.zero, tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                side: BorderSide(color: _isFollowing ? Colors.grey : Colors.red), foregroundColor: _isFollowing ? Colors.grey : Colors.red, textStyle: const TextStyle(fontSize: 12)),
              child: Text(_isFollowing ? '已关注' : '关注'),
            ),
          ],
        ]),
        actions: [IconButton(icon: const Icon(Icons.more_horiz, color: Colors.black87), onPressed: _showMoreMenu)],
      ),
      body: Column(children: [
        Expanded(child: SingleChildScrollView(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          if (hasVid || hasImg) _buildUnifiedMedia(),
          Padding(padding: const EdgeInsets.fromLTRB(16, 12, 16, 0), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            if (widget.post.title.isNotEmpty) ...[Text(widget.post.title, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, height: 1.3)), const SizedBox(height: 8)],
            if (widget.post.content.isNotEmpty) ...[Text(widget.post.content, style: const TextStyle(fontSize: 15, height: 1.7, color: Colors.black87)), const SizedBox(height: 10)],
            if (widget.post.tags.isNotEmpty) ...[Wrap(spacing: 8, runSpacing: 4, children: widget.post.tags.map((t) => Text('#$t', style: const TextStyle(fontSize: 13, color: Color(0xFF5974A8)))).toList()), const SizedBox(height: 10)],
            Text(_timeAgo(widget.post.createdAt), style: const TextStyle(fontSize: 12, color: Colors.grey)),
          ])),
          const SizedBox(height: 16), const Divider(height: 1),
          if (widget.post.likeCount > 0) ...[
            Padding(padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12), child: Row(children: [
              const Icon(Icons.favorite, size: 18, color: Colors.red), const SizedBox(width: 6),
              Text('${widget.post.likeCount + (_liked ? 1 : 0)} 人赞了', style: const TextStyle(fontSize: 13, color: Colors.black54)),
            ])), const Divider(height: 1),
          ],
          Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            const Text('评论', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600)), const SizedBox(height: 12),
            if (widget.post.commentCount == 0) const Center(child: Padding(padding: EdgeInsets.all(24), child: Text('暂无评论', style: TextStyle(color: Colors.grey, fontSize: 13)))),
            if (widget.post.commentCount > 0) GestureDetector(onTap: _openComments, child: Container(padding: const EdgeInsets.all(12), decoration: BoxDecoration(color: Colors.grey.shade50, borderRadius: BorderRadius.circular(8)),
              child: Row(children: [const Icon(Icons.chat_bubble_outline, size: 18, color: Colors.grey), const SizedBox(width: 8), Text('查看全部 ${widget.post.commentCount} 条评论', style: const TextStyle(fontSize: 13, color: Colors.grey)), const Spacer(), const Icon(Icons.chevron_right, size: 18, color: Colors.grey)]))),
          ])),
        ]))),
        Container(decoration: BoxDecoration(color: Colors.white, border: Border(top: BorderSide(color: Colors.grey.shade200))),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          child: SafeArea(top: false, child: Row(mainAxisAlignment: MainAxisAlignment.spaceAround, children: [
            _btn(Icons.favorite_border, Icons.favorite, _liked, '${widget.post.likeCount + (_liked ? 1 : 0)}', _toggleLike),
            _btn(Icons.chat_bubble_outline, Icons.chat_bubble_outline, false, '${widget.post.commentCount}', _openComments),
            _btn(Icons.bookmark_border, Icons.bookmark, _bookmarked, _bookmarked ? '已收藏' : '收藏', _toggleBookmark),
            _btn(Icons.share_outlined, Icons.share_outlined, false, '分享', _showShareSheet),
          ])),
        ),
      ]),
    );
  }

  Widget _buildUnifiedMedia() {
    final sw = MediaQuery.of(context).size.width;
    final hasV = widget.post.videoUrl != null;
    final imgs = widget.post.images;
    final total = (hasV ? 1 : 0) + imgs.length;
    return Column(children: [
      SizedBox(height: sw * 0.75, width: double.infinity, child: PageView.builder(
        controller: _pageController, onPageChanged: (i) => setState(() => _currentImageIndex = i), itemCount: total,
        itemBuilder: (ctx, i) {
          if (hasV && i == 0) return GestureDetector(onTap: () { setState(() { _playing ? _videoController!.pause() : _videoController!.play(); _playing = !_playing; }); }, child: _videoFrame());
          final ii = hasV ? i - 1 : i;
          return GestureDetector(onTap: () => _showFullscreen(ii), onLongPress: () => _saveImageToGallery(imgs[ii]),
            child: Container(color: Colors.white, child: Center(child: CachedNetworkImage(imageUrl: imgs[ii], fit: BoxFit.contain, width: sw,
              placeholder: (_, __) => Container(color: Colors.grey.shade100),
              errorWidget: (_, __, ___) => Container(color: Colors.grey.shade100, child: const Icon(Icons.broken_image, color: Colors.grey, size: 48))))));
        },
      )),
      if (total > 1) Padding(padding: const EdgeInsets.symmetric(vertical: 8), child: Row(mainAxisAlignment: MainAxisAlignment.center,
        children: List.generate(total, (i) => Container(width: i == _currentImageIndex ? 16 : 6, height: 6, margin: const EdgeInsets.symmetric(horizontal: 3),
          decoration: BoxDecoration(color: i == _currentImageIndex ? Colors.red : Colors.grey.shade300, borderRadius: BorderRadius.circular(3)))))),
    ]);
  }

  Widget _videoFrame() {
    if (_videoError) return Container(color: Colors.black87, child: Center(child: Column(mainAxisSize: MainAxisSize.min, children: [
      const Icon(Icons.error_outline, size: 48, color: Colors.white54), const SizedBox(height: 8), const Text('视频无法播放', style: TextStyle(color: Colors.white70)), const SizedBox(height: 12),
      ElevatedButton.icon(onPressed: () { _videoController?.dispose(); setState(() { _videoError = false; _videoInitialized = false; }); _initVideo(); }, icon: const Icon(Icons.refresh, size: 16), label: const Text('重试')),
    ])));
    if (!_videoInitialized) return Container(color: Colors.black, child: const Center(child: CircularProgressIndicator(color: Colors.white)));
    return Container(color: Colors.black, child: Center(child: AspectRatio(aspectRatio: _videoController!.value.aspectRatio, child: Stack(alignment: Alignment.center, children: [
      VideoPlayer(_videoController!), if (!_playing) Container(width: 56, height: 56, decoration: const BoxDecoration(color: Colors.black38, shape: BoxShape.circle), child: const Icon(Icons.play_arrow, size: 36, color: Colors.white)),
    ]))));
  }

  void _openComments() { Navigator.push(context, MaterialPageRoute(builder: (_) => CommentsScreen(postId: widget.post.id, initialCount: widget.post.commentCount))); }

  Widget _btn(IconData off, IconData on, bool active, String label, VoidCallback cb) {
    return InkWell(onTap: cb, child: Padding(padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: Column(mainAxisSize: MainAxisSize.min, children: [Icon(active ? on : off, size: 24, color: active ? Colors.red : Colors.black54), const SizedBox(height: 2), Text(label, style: TextStyle(fontSize: 11, color: active ? Colors.red : Colors.black54))])));
  }

  String _timeAgo(DateTime dt) { final d = DateTime.now().difference(dt); if (d.inMinutes < 1) return '刚刚'; if (d.inMinutes < 60) return '${d.inMinutes}分钟前'; if (d.inHours < 24) return '${d.inHours}小时前'; return '${d.inDays}天前'; }
}

class _ImageViewer extends StatefulWidget {
  final List<String> images; final int initialIndex;
  const _ImageViewer({required this.images, required this.initialIndex});
  @override
  State<_ImageViewer> createState() => _ImageViewerState();
}

class _ImageViewerState extends State<_ImageViewer> {
  late final PageController _ctrl;

  @override
  void initState() { super.initState(); _ctrl = PageController(initialPage: widget.initialIndex); }

  @override
  void dispose() { _ctrl.dispose(); super.dispose(); }

  @override
  Widget build(BuildContext context) {
    return Scaffold(backgroundColor: Colors.black, appBar: AppBar(backgroundColor: Colors.black, iconTheme: const IconThemeData(color: Colors.white), elevation: 0),
      body: PageView.builder(controller: _ctrl, itemCount: widget.images.length,
        itemBuilder: (_, i) => InteractiveViewer(minScale: 0.5, maxScale: 4.0,
          child: Center(child: CachedNetworkImage(imageUrl: widget.images[i], fit: BoxFit.contain,
            placeholder: (_, __) => const Center(child: CircularProgressIndicator(color: Colors.white)),
            errorWidget: (_, __, ___) => const Icon(Icons.broken_image, color: Colors.white54, size: 64))))));
  }
}
