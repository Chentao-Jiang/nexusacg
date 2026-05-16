import 'package:flutter/material.dart';
import 'package:nexusacg/presentation/screens/profile/edit_profile_screen.dart';

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('设置')),
      body: Column(
        children: [
          _settingsGroup('账户', [
            _settingsTile(context, '编辑个人资料', Icons.person_outline, () {
              Navigator.push(context, MaterialPageRoute(builder: (_) => const EditProfileScreen()));
            }),
            _settingsTile(context, '收货地址', Icons.location_on_outlined, null),
            _settingsTile(context, '消息通知', Icons.notifications_outlined, null),
          ]),
          _settingsGroup('通用', [
            _settingsTile(context, '清除缓存', Icons.delete_outline, null),
            _settingsTile(context, '关于我们', Icons.info_outline, null),
            _settingsTile(context, '用户协议', Icons.description_outlined, null),
            _settingsTile(context, '隐私政策', Icons.security_outlined, null),
          ]),
        ],
      ),
    );
  }

  Widget _settingsGroup(String title, List<Widget> children) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Text(title, style: const TextStyle(fontSize: 13, color: Colors.grey)),
        ),
        Card(
          margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
          child: Column(children: children),
        ),
      ],
    );
  }

  Widget _settingsTile(BuildContext context, String title, IconData icon, VoidCallback? onTap) {
    return ListTile(
      leading: Icon(icon, size: 22),
      title: Text(title),
      trailing: const Icon(Icons.chevron_right),
      onTap: onTap ?? () {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('$title 功能开发中')),
        );
      },
    );
  }
}
