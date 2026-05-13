import 'package:flutter/material.dart';
import 'package:nexusacg/core/models/models.dart';

class EventsScreen extends StatefulWidget {
  const EventsScreen({super.key});

  @override
  State<EventsScreen> createState() => _EventsScreenState();
}

class _EventsScreenState extends State<EventsScreen> {
  final List<EventModel> _events = []; // Placeholder
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadEvents();
  }

  Future<void> _loadEvents() async {
    // TODO: Load from API
    await Future.delayed(const Duration(seconds: 1));
    setState(() => _loading = false);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('漫展活动')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _loadEvents,
              child: _events.isEmpty
                  ? ListView(
                      children: [
                        Container(
                          height: 200,
                          margin: const EdgeInsets.all(24),
                          decoration: BoxDecoration(
                            color: Colors.grey.shade100,
                            borderRadius: BorderRadius.circular(16),
                          ),
                          child: const Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.event, size: 48, color: Colors.grey),
                              SizedBox(height: 12),
                              Text('暂无漫展活动', style: TextStyle(color: Colors.grey)),
                              SizedBox(height: 4),
                              Text('快来添加第一个活动吧', style: TextStyle(color: Colors.grey, fontSize: 12)),
                            ],
                          ),
                        ),
                      ],
                    )
                  : ListView.builder(
                      padding: const EdgeInsets.all(12),
                      itemCount: _events.length,
                      itemBuilder: (context, index) => _EventCard(_events[index]),
                    ),
            ),
    );
  }
}

class _EventCard extends StatelessWidget {
  final EventModel event;
  const _EventCard(this.event);

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: Theme.of(context).primaryColor.withOpacity(0.1),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(Icons.event, color: Theme.of(context).primaryColor),
        ),
        title: Text(event.name, style: const TextStyle(fontWeight: FontWeight.bold)),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            Text('${event.address}'),
            Text('${event.startTime.month}月${event.startTime.day}日'),
          ],
        ),
        trailing: const Icon(Icons.chevron_right),
        onTap: () {},
      ),
    );
  }
}
