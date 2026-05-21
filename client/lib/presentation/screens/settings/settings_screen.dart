import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/presentation/screens/profile/edit_profile_screen.dart';
import 'package:nexusacg/presentation/screens/settings/legal_pages.dart';
import 'package:nexusacg/presentation/screens/settings/address_screen.dart';

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
            _settingsTile(context, '收货地址', Icons.location_on_outlined, () {
                Navigator.push(context, MaterialPageRoute(builder: (_) => const AddressListScreen()));
              }),
            _settingsTile(context, '消息通知', Icons.notifications_outlined, null),
          ]),
          _settingsGroup('通用', [
            _settingsTile(context, '清除缓存', Icons.delete_outline, () async {
                await _clearCache(context);
              }),
            _settingsTile(context, '关于我们', Icons.info_outline, () {
                _showAbout(context);
              }),
            _settingsTile(context, '用户协议', Icons.description_outlined, () {
                Navigator.push(context, MaterialPageRoute(builder: (_) => const LegalPage(
                  title: '用户协议',
                  content: userAgreement,
                )));
              }),
            _settingsTile(context, '隐私政策', Icons.security_outlined, () {
                Navigator.push(context, MaterialPageRoute(builder: (_) => const LegalPage(
                  title: '隐私政策',
                  content: privacyPolicy,
                )));
              }),
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

  static Future<void> _clearCache(BuildContext context) async {
    await CachedNetworkImage.evictFromCache('');
    try { PaintingBinding.instance.imageCache.clear(); } catch (_) {}
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('缓存已清除')));
    }
  }

  static void _showAbout(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('关于次元链'),
        content: const Text('次元链 NexusACG v0.1.0\n\nACG线下产业生态服务平台\n\n为ACG爱好者提供一站式服务：\n· 妆娘/摄影/摊位服务\n· 漫展活动报名\n· 社区交流分享\n· 二手商品交易'),
        actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('确定'))],
      ),
    );
  }
}