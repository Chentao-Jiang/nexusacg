import 'package:flutter/material.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/screens/orders/order_detail_screen.dart';

class CheckoutScreen extends StatefulWidget {
  final String productId;
  final int quantity;
  final ProductModel product;
  const CheckoutScreen({super.key, required this.productId, required this.quantity, required this.product});

  @override
  State<CheckoutScreen> createState() => _CheckoutScreenState();
}

class _CheckoutScreenState extends State<CheckoutScreen> {
  final _repo = OrderRepository();
  bool _submitting = false;
  String _selectedPayment = 'alipay';

  Future<void> _createOrder() async {
    setState(() => _submitting = true);
    try {
      final result = await _repo.createOrder([
        (productId: widget.productId, quantity: widget.quantity),
      ]);
      if (result['code'] == 0 && result['data'] != null) {
        final orderNo = result['data']['order_no'] as String;
        if (mounted) {
          // Go directly to order detail where pay button is available
          Navigator.of(context).pop(); // Pop checkout
          Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => OrderDetailScreen(orderNo: orderNo)),
          );
        }
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(result['message'] as String? ?? '创建订单失败')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('创建订单失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  double get _total => widget.product.price * widget.quantity;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(title: const Text('确认订单')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Product info
            _sectionTitle('商品信息'),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    Container(
                      width: 80,
                      height: 80,
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(8),
                        color: Colors.grey.shade200,
                      ),
                      child: widget.product.images.isNotEmpty
                          ? ClipRRect(
                              borderRadius: BorderRadius.circular(8),
                              child: Image.network(widget.product.images.first, fit: BoxFit.cover,
                                  errorBuilder: (_, __, ___) => const Icon(Icons.image, size: 40, color: Colors.grey)),
                            )
                          : const Icon(Icons.image, size: 40, color: Colors.grey),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(widget.product.name,
                              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                              maxLines: 2, overflow: TextOverflow.ellipsis),
                          const SizedBox(height: 8),
                          Text('¥${widget.product.price.toStringAsFixed(2)}',
                              style: TextStyle(color: theme.colorScheme.error, fontWeight: FontWeight.bold, fontSize: 16)),
                          const SizedBox(height: 4),
                          Text('数量: ${widget.quantity}', style: const TextStyle(color: Colors.grey)),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 20),
            // Payment method
            _sectionTitle('支付方式'),
            Card(
              child: Column(
                children: [
                  _paymentOption('alipay', '支付宝', Icons.account_balance_wallet),
                  const Divider(height: 1),
                  _paymentOption('wechat', '微信支付', Icons.wechat_outlined),
                ],
              ),
            ),
          ],
        ),
      ),
      bottomNavigationBar: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: theme.scaffoldBackgroundColor,
          boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 8)],
        ),
        child: SafeArea(
          child: Row(
            children: [
              Text('合计: ', style: const TextStyle(fontSize: 14)),
              Text('¥${_total.toStringAsFixed(2)}',
                  style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: theme.colorScheme.error)),
              const Spacer(),
              FilledButton(
                onPressed: _submitting ? null : _createOrder,
                style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 14)),
                child: _submitting
                    ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                    : const Text('提交订单'),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _sectionTitle(String title) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
    );
  }

  Widget _paymentOption(String value, String label, IconData icon) {
    final selected = _selectedPayment == value;
    return RadioListTile(
      value: value,
      groupValue: _selectedPayment,
      onChanged: (v) => setState(() => _selectedPayment = v.toString()),
      title: Row(
        children: [
          Icon(icon, size: 22),
          const SizedBox(width: 8),
          Text(label, style: const TextStyle(fontSize: 15)),
        ],
      ),
      secondary: selected ? Icon(Icons.check_circle, color: Theme.of(context).colorScheme.primary) : null,
    );
  }
}
