import 'package:flutter/material.dart';
import 'package:nexusacg/core/network/api_client.dart';

class MyRegistrationsScreen extends StatefulWidget {
  const MyRegistrationsScreen({super.key});

  @override
  State<MyRegistrationsScreen> createState() => _MyRegistrationsScreenState();
}

class _MyRegistrationsScreenState extends State<MyRegistrationsScreen> {
  List<Map<String, dynamic>> _events = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final res = await ApiClient().get('/events/my-registrations');
      final d = res.data;
      if (d is Map && d['code'] == 0 && d['data'] != null) {
        final items = (d['data'] as Map)['items'] as List? ?? [];
        setState(() {
          _events = items.cast<Map<String, dynamic>>();
          _loading = false;
        });
      } else {
        setState(() => _loading = false);
      }
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('我的预约')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _events.isEmpty
              ? const Center(child: Text('暂无预约'))
              : ListView.builder(
                  padding: const EdgeInsets.all(12),
                  itemCount: _events.length,
                  itemBuilder: (_, i) {
                    final e = _events[i];
                    return Card(
                      margin: const EdgeInsets.only(bottom: 10),
                      child: ListTile(
                        leading: const Icon(Icons.event, size: 32),
                        title: Text(e['name'] ?? '', maxLines: 1, overflow: TextOverflow.ellipsis),
                        subtitle: Text(e['start_time']?.toString().substring(0, 10) ?? ''),
                        trailing: const Icon(Icons.check_circle, color: Colors.green, size: 20),
                      ),
                    );
                  },
                ),
    );
  }
}
