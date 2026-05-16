import 'package:flutter/material.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';

class PaymentScreen extends StatefulWidget {
  final OrderModel order;
  const PaymentScreen({super.key, required this.order});

  @override
  State<PaymentScreen> createState() => _PaymentScreenState();
}

class _PaymentScreenState extends State<PaymentScreen> {
  final _repo = OrderRepository();
  String _selectedMethod = 'alipay';
  bool _processing = false;

  Future<void> _pay() async {
    setState(() => _processing = true);
    try {
      // Generate a client-side payment ID (in real app, this comes from payment SDK)
      final paymentId = DateTime.now().millisecondsSinceEpoch.toString();
      final result = await _repo.payOrder(
        widget.order.orderNo,
        paymentMethod: _selectedMethod,
        paymentId: paymentId,
      );
      if (mounted) {
        if (result['code'] == 0) {
          Navigator.pop(context, true); // success
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('支付成功')));
        } else {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(result['message'] as String? ?? '支付失败')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('支付失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _processing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(title: const Text('选择支付方式')),
      body: Column(
        children: [
          const SizedBox(height: 24),
          Text('应付金额', style: TextStyle(fontSize: 16, color: Colors.grey.shade600)),
          const SizedBox(height: 8),
          Text('¥${widget.order.totalAmount.toStringAsFixed(2)}',
              style: TextStyle(fontSize: 32, fontWeight: FontWeight.bold, color: theme.colorScheme.error)),
          const SizedBox(height: 32),
          // Payment methods
          Card(
            margin: const EdgeInsets.symmetric(horizontal: 16),
            child: Column(
              children: [
                _paymentOption('alipay', '支付宝', Icons.account_balance_wallet, Icons.payment),
                const Divider(height: 1),
                _paymentOption('wechat', '微信支付', Icons.wechat_outlined, Icons.payment),
              ],
            ),
          ),
          const Spacer(),
          // Pay button
          Container(
            padding: const EdgeInsets.all(16),
            width: double.infinity,
            child: FilledButton(
              onPressed: _processing ? null : _pay,
              style: FilledButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              ),
              child: _processing
                  ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                  : const Text('确认支付', style: TextStyle(fontSize: 16)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _paymentOption(String value, String label, IconData icon, IconData subIcon) {
    final selected = _selectedMethod == value;
    final theme = Theme.of(context);
    return RadioListTile(
      value: value,
      groupValue: _selectedMethod,
      onChanged: (v) => setState(() => _selectedMethod = v.toString()),
      title: Row(
        children: [
          Icon(icon, size: 28, color: value == 'alipay' ? const Color(0xFF1677FF) : const Color(0xFF07C160)),
          const SizedBox(width: 12),
          Text(label, style: const TextStyle(fontSize: 16)),
        ],
      ),
      secondary: selected
          ? Icon(Icons.check_circle, color: theme.colorScheme.primary)
          : null,
    );
  }
}
