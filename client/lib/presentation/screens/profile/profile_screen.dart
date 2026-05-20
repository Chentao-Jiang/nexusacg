import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_bloc.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_state.dart';
import 'package:nexusacg/presentation/screens/orders/orders_screen.dart';
import 'package:nexusacg/presentation/screens/settings/settings_screen.dart';
import 'package:nexusacg/presentation/screens/community/my_posts_screen.dart';
import 'package:nexusacg/presentation/screens/community/follow_list_screen.dart';
import 'package:nexusacg/presentation/screens/profile/my_registrations_screen.dart';
import 'package:nexusacg/presentation/screens/profile/edit_profile_screen.dart';
import 'package:nexusacg/presentation/screens/certification/certification_screen.dart';

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: BlocBuilder<AuthBloc, AuthState>(
          builder: (context, state) {
            final user = state is AuthAuthenticated ? state.user : null;
            return ListView(
              children: [
                // Profile header
                GestureDetector(
                  onTap: () {
                    Navigator.push(context, MaterialPageRoute(builder: (_) => const EditProfileScreen()));
                  },
                  child: Container(
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [Color(0xFF6366F1), Color(0xFF8B5CF6)],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                    ),
                    child: Column(
                      children: [
                        CircleAvatar(
                          radius: 40,
                          backgroundImage: user?.avatarUrl != null
                              ? CachedNetworkImageProvider(user!.avatarUrl!)
                              : null,
                          child: user?.avatarUrl == null
                              ? const Icon(Icons.person, size: 40, color: Colors.white)
                              : null,
                        ),
                        const SizedBox(height: 12),
                        Text(
                          user?.nickname ?? '用户',
                          style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.white),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          user != null && user.bio.isNotEmpty ? user.bio : 'ACG爱好者',
                          style: const TextStyle(fontSize: 14, color: Colors.white70),
                        ),
                      ],
                    ),
                  ),
                ),
                // Stats
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      _statItem('帖子', '0'),
                      _statItem('关注', '0'),
                      _statItem('粉丝', '0'),
                    ],
                  ),
                ),
                const Divider(),
                // Quick actions - order status shortcuts
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceAround,
                    children: [
                      _quickAction(context, Icons.pending_actions, '待付款', 'pending'),
                      _quickAction(context, Icons.local_shipping, '待收货', 'shipped'),
                      _quickAction(context, Icons.check_circle, '已完成', 'completed'),
                      _quickAction(context, Icons.receipt_long, '退款', 'refunded'),
                    ],
                  ),
                ),
                const Divider(),
                // Menu items
                _menuItem(Icons.shopping_bag, '我的订单', onTap: () {
                  Navigator.push(context, MaterialPageRoute(builder: (_) => const OrdersScreen()));
                }),
                _menuItem(Icons.article_outlined, '我的帖子', onTap: () {
                  Navigator.push(context, MaterialPageRoute(builder: (_) => const MyPostsScreen()));
                }),
                _menuItem(Icons.favorite, '我的收藏', onTap: () => _showComingSoon(context, '我的收藏')),
                _menuItem(Icons.local_offer, '我的商品', onTap: () => _showComingSoon(context, '我的商品')),
                _menuItem(Icons.bookmark, '我的预约', onTap: () {
                  Navigator.push(context, MaterialPageRoute(builder: (_) => const MyRegistrationsScreen()));
                }),
                const Divider(),
                _menuItem(Icons.store, '我要入驻', subtitle: '妆娘/摄影师/摊主', onTap: () {
                  Navigator.push(context, MaterialPageRoute(builder: (_) => const CertificationScreen()));
                }),
                _menuItem(Icons.settings, '设置', onTap: () {
                  Navigator.push(context, MaterialPageRoute(builder: (_) => const SettingsScreen()));
                }),
                _menuItem(Icons.help_outline, '帮助与反馈', onTap: () {
                  Navigator.push(context, MaterialPageRoute(builder: (_) => const _HelpScreen()));
                }),
                const Divider(),
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: OutlinedButton(
                    onPressed: () {
                      context.read<AuthBloc>().add(AuthLogoutRequested());
                    },
                    style: OutlinedButton.styleFrom(
                      foregroundColor: Theme.of(context).colorScheme.error,
                      side: BorderSide(color: Theme.of(context).colorScheme.error),
                    ),
                    child: const Text('退出登录'),
                  ),
                ),
                const SizedBox(height: 16),
                const Center(child: Text('次元链 v0.1.0', style: TextStyle(color: Colors.grey, fontSize: 12))),
                const SizedBox(height: 24),
              ],
            );
          },
        ),
      ),
    );
  }

  Widget _statItem(String label, String value) {
    return Expanded(
      child: Column(
        children: [
          Text(value, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
          const SizedBox(height: 4),
          Text(label, style: const TextStyle(color: Colors.grey, fontSize: 13)),
        ],
      ),
    );
  }

  Widget _quickAction(BuildContext context, IconData icon, String label, String status) {
    return InkWell(
      onTap: () {
        Navigator.push(context, MaterialPageRoute(builder: (_) => OrdersScreen(initialStatus: status)));
      },
      child: Column(
        children: [
          Icon(icon, size: 28),
          const SizedBox(height: 4),
          Text(label, style: const TextStyle(fontSize: 12)),
        ],
      ),
    );
  }

  void _showComingSoon(BuildContext context, String feature) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('$feature 功能开发中')),
    );
  }

  Widget _menuItem(IconData icon, String title, {String? subtitle, VoidCallback? onTap}) {
    return ListTile(
      leading: Icon(icon),
      title: Text(title),
      subtitle: subtitle != null ? Text(subtitle, style: const TextStyle(fontSize: 12, color: Colors.grey)) : null,
      trailing: const Icon(Icons.chevron_right, size: 16),
      onTap: onTap ?? () {},
    );
  }
}


class _HelpScreen extends StatelessWidget {
  const _HelpScreen();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('帮助与反馈')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          ListTile(
            leading: const Icon(Icons.question_answer, color: Colors.blue),
            title: const Text('常见问题'),
            subtitle: const Text('查看常见问题解答'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () {},
          ),
          ListTile(
            leading: const Icon(Icons.mail_outline, color: Colors.green),
            title: const Text('联系客服'),
            subtitle: const Text('发送邮件至 support@nexusacg.com'),
            onTap: () {},
          ),
          ListTile(
            leading: const Icon(Icons.feedback_outlined, color: Colors.orange),
            title: const Text('意见反馈'),
            subtitle: const Text('告诉我们你的想法'),
            onTap: () {},
          ),
          const Divider(),
          const Padding(
            padding: EdgeInsets.all(16),
            child: Text('次元链 NexusACG v0.1.0', style: TextStyle(color: Colors.grey, fontSize: 12)),
          ),
        ],
      ),
    );
  }
}

