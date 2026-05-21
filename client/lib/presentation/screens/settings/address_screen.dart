import 'package:flutter/material.dart';
import 'package:nexusacg/core/network/api_client.dart';

class AddressListScreen extends StatefulWidget {
  const AddressListScreen({super.key});
  @override
  State<AddressListScreen> createState() => _AddressListScreenState();
}

class _AddressListScreenState extends State<AddressListScreen> {
  List<Map<String, dynamic>> _addresses = [];
  bool _loading = true;

  @override
  void initState() { super.initState(); _load(); }

  Future<void> _load() async {
    final res = await ApiClient().get('/addresses');
    final d = res.data;
    if (d is Map && d['code'] == 0 && d['data'] != null) {
      final items = (d['data'] as Map)['items'] as List? ?? [];
      setState(() { _addresses = items.cast<Map<String, dynamic>>(); _loading = false; });
    } else {
      setState(() => _loading = false);
    }
  }

  Future<void> _delete(String id) async {
    await ApiClient().delete('/addresses/$id');
    _load();
  }

  Future<void> _edit(Map<String, dynamic>? addr) async {
    final result = await Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => AddressEditScreen(address: addr)),
    );
    if (result == true) _load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('收货地址')),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _edit(null),
        child: const Icon(Icons.add),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _addresses.isEmpty
              ? const Center(child: Text('暂无收货地址'))
              : ListView.builder(
                  padding: const EdgeInsets.all(12),
                  itemCount: _addresses.length,
                  itemBuilder: (_, i) {
                    final a = _addresses[i];
                    return Card(
                      margin: const EdgeInsets.only(bottom: 10),
                      child: ListTile(
                        title: Row(
                          children: [
                            Text(a['name'] ?? '', style: const TextStyle(fontWeight: FontWeight.w500)),
                            if (a['is_default'] == true) ...[
                              const SizedBox(width: 8),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                                decoration: BoxDecoration(color: Colors.red.shade50, borderRadius: BorderRadius.circular(3)),
                                child: const Text('默认', style: TextStyle(color: Colors.red, fontSize: 10)),
                              ),
                            ],
                          ],
                        ),
                        subtitle: Text('${a['province'] ?? ''} ${a['city'] ?? ''} ${a['district'] ?? ''} ${a['detail'] ?? ''}\n${a['phone'] ?? ''}'),
                        trailing: PopupMenuButton(
                          itemBuilder: (_) => [
                            const PopupMenuItem(value: 'edit', child: Text('编辑')),
                            const PopupMenuItem(value: 'delete', child: Text('删除', style: TextStyle(color: Colors.red))),
                          ],
                          onSelected: (v) {
                            if (v == 'edit') _edit(a);
                            if (v == 'delete') _delete(a['id']?.toString() ?? '');
                          },
                        ),
                      ),
                    );
                  },
                ),
    );
  }
}

class AddressEditScreen extends StatefulWidget {
  final Map<String, dynamic>? address;
  const AddressEditScreen({super.key, this.address});
  @override
  State<AddressEditScreen> createState() => _AddressEditScreenState();
}

class _AddressEditScreenState extends State<AddressEditScreen> {
  final _name = TextEditingController();
  final _phone = TextEditingController();
  final _detail = TextEditingController();
  final _province = TextEditingController();
  final _city = TextEditingController();
  final _district = TextEditingController();
  bool _isDefault = false;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final a = widget.address;
    if (a != null) {
      _name.text = a['name'] ?? '';
      _phone.text = a['phone'] ?? '';
      _detail.text = a['detail'] ?? '';
      _province.text = a['province'] ?? '';
      _city.text = a['city'] ?? '';
      _district.text = a['district'] ?? '';
      _isDefault = a['is_default'] == true;
    }
  }

  Future<void> _save() async {
    if (_name.text.isEmpty || _phone.text.isEmpty || _detail.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请填写完整信息')));
      return;
    }
    setState(() => _saving = true);
    final body = {
      'name': _name.text,
      'phone': _phone.text,
      'province': _province.text,
      'city': _city.text,
      'district': _district.text,
      'detail': _detail.text,
      'is_default': _isDefault,
    };
    final a = widget.address;
    final res = a != null
        ? await ApiClient().put('/addresses/${a['id']}', data: body)
        : await ApiClient().post('/addresses', data: body);
    setState(() => _saving = false);
    if (mounted && res.data is Map && (res.data as Map)['code'] == 0) {
      Navigator.pop(context, true);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(widget.address != null ? '编辑地址' : '新增地址')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            TextField(controller: _name, decoration: const InputDecoration(labelText: '收货人', border: OutlineInputBorder())),
            const SizedBox(height: 12),
            TextField(controller: _phone, decoration: const InputDecoration(labelText: '手机号', border: OutlineInputBorder()), keyboardType: TextInputType.phone),
            const SizedBox(height: 12),
            Row(children: [
              Expanded(child: TextField(controller: _province, decoration: const InputDecoration(labelText: '省', border: OutlineInputBorder()))),
              const SizedBox(width: 8),
              Expanded(child: TextField(controller: _city, decoration: const InputDecoration(labelText: '市', border: OutlineInputBorder()))),
              const SizedBox(width: 8),
              Expanded(child: TextField(controller: _district, decoration: const InputDecoration(labelText: '区', border: OutlineInputBorder()))),
            ]),
            const SizedBox(height: 12),
            TextField(controller: _detail, decoration: const InputDecoration(labelText: '详细地址', border: OutlineInputBorder()), maxLines: 2),
            const SizedBox(height: 12),
            SwitchListTile(title: const Text('设为默认地址'), value: _isDefault, onChanged: (v) => setState(() => _isDefault = v), contentPadding: EdgeInsets.zero),
            const SizedBox(height: 20),
            SizedBox(width: double.infinity, height: 48,
              child: FilledButton(onPressed: _saving ? null : _save, child: const Text('保存'))),
          ],
        ),
      ),
    );
  }
}
