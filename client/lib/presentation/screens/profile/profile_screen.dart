import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_bloc.dart';

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: ListView(
          children: [
            // Profile header
            Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    Theme.of(context).primaryColor,
                    Theme.of(context).primaryColor.withOpacity(0.7),
                  ],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
              ),
              child: const Column(
                children: [
                  CircleAvatar(radius: 40, child: Icon(Icons.person, size: 40)),
                  SizedBox(height: 12),
                  Text('用户昵称', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.white)),
                  SizedBox(height: 4),
                  Text('ACG爱好者', style: TextStyle(fontSize: 14, color: Colors.white70)),
                ],
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
            // Menu items
            _menuItem(Icons.shopping_bag, '我的订单'),
            _menuItem(Icons.favorite, '我的收藏'),
            _menuItem(Icons.local_offer, '我的商品'),
            _menuItem(Icons.bookmark, '我的预约'),
            const Divider(),
            _menuItem(Icons.store, '我要入驻', subtitle: '妆娘/摄影师/摊主'),
            _menuItem(Icons.settings, '设置'),
            _menuItem(Icons.help_outline, '帮助与反馈'),
            const Divider(),
            Padding(
              padding: const EdgeInsets.all(16),
              child: OutlinedButton(
                onPressed: () {
                  context.read<AuthBloc>().add(AuthLogoutRequested());
                },
                child: const Text('退出登录'),
              ),
            ),
            const SizedBox(height: 16),
            const Center(child: Text('次元链 v0.1.0', style: TextStyle(color: Colors.grey, fontSize: 12))),
            const SizedBox(height: 24),
          ],
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

  Widget _menuItem(IconData icon, String title, {String? subtitle}) {
    return ListTile(
      leading: Icon(icon),
      title: Text(title),
      subtitle: subtitle != null ? Text(subtitle, style: const TextStyle(fontSize: 12, color: Colors.grey)) : null,
      trailing: const Icon(Icons.chevron_right, size: 16),
      onTap: () {},
    );
  }
}
