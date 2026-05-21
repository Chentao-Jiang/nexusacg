import 'package:flutter/material.dart';
import 'package:nexusacg/core/network/api_client.dart';

class ChatScreen extends StatefulWidget {
  final String otherUserId;
  final String otherName;
  const ChatScreen({super.key, required this.otherUserId, required this.otherName});
  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final _msgCtrl = TextEditingController();
  List<Map<String, dynamic>> _messages = [];
  bool _loading = true;

  @override
  void initState() { super.initState(); _load(); }

  Future<void> _load() async {
    final res = await ApiClient().get('/messages/${widget.otherUserId}');
    final d = res.data;
    if (d is Map && d['code'] == 0 && d['data'] != null) {
      final data = d['data'] as Map;
      setState(() {
        _messages = (data['items'] as List?)?.cast<Map<String, dynamic>>().reversed.toList() ?? [];
        _loading = false;
      });
    } else {
      setState(() => _loading = false);
    }
  }

  Future<void> _send() async {
    final text = _msgCtrl.text.trim();
    if (text.isEmpty) return;
    _msgCtrl.clear();
    await ApiClient().post('/messages', data: {'receiver_id': widget.otherUserId, 'content': text});
    _load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(widget.otherName)),
      body: Column(children: [
        Expanded(
          child: _loading
              ? const Center(child: CircularProgressIndicator())
              : _messages.isEmpty
                  ? const Center(child: Text('发送第一条消息'))
                  : ListView.builder(
                      reverse: true,
                      padding: const EdgeInsets.all(12),
                      itemCount: _messages.length,
                      itemBuilder: (_, i) {
                        final m = _messages[i];
                        final isMe = m['receiver_id'] == widget.otherUserId;
                        return Align(
                          alignment: isMe ? Alignment.centerRight : Alignment.centerLeft,
                          child: Container(
                            margin: const EdgeInsets.only(bottom: 8),
                            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                            constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.7),
                            decoration: BoxDecoration(
                              color: isMe ? Theme.of(context).primaryColor : Colors.grey.shade200,
                              borderRadius: BorderRadius.only(
                                topLeft: const Radius.circular(16),
                                topRight: const Radius.circular(16),
                                bottomLeft: Radius.circular(isMe ? 16 : 4),
                                bottomRight: Radius.circular(isMe ? 4 : 16),
                              ),
                            ),
                            child: Text(m['content'] ?? '', style: TextStyle(fontSize: 15, color: isMe ? Colors.white : Colors.black87)),
                          ),
                        );
                      },
                    ),
        ),
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(color: Colors.white, boxShadow: [BoxShadow(color: Colors.black12, blurRadius: 4, offset: const Offset(0, -2))]),
          child: SafeArea(top: false, child: Row(children: [
            Expanded(child: TextField(controller: _msgCtrl, decoration: const InputDecoration(hintText: '输入消息...', border: OutlineInputBorder(), contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 8)))),
            const SizedBox(width: 8),
            IconButton(icon: const Icon(Icons.send, color: Colors.blue), onPressed: _send),
          ])),
        ),
      ]),
    );
  }
}
