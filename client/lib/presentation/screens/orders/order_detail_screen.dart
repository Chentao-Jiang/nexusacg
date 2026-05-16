import 'package:flutter/material.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/screens/orders/payment_screen.dart';

class OrderDetailScreen extends StatefulWidget {
  final String orderNo;
  const OrderDetailScreen({super.key, required this.orderNo});

  @override
  State<OrderDetailScreen> createState() => _OrderDetailScreenState();
}

class _OrderDetailScreenState extends State<OrderDetailScreen> {
  final _repo = OrderRepository();
  OrderModel? _order;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadOrder();
  }

  Future<void> _loadOrder() async {
    setState(() => _loading = true);
    try {
      final order = await _repo.getOrder(widget.orderNo);
      if (mounted) setState(() => _order = order);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加载失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _cancelOrder() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('取消订单'),
        content: const Text('确定要取消此订单吗？'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('取消')),
          TextButton(onPressed: () => Navigator.pop(context, true), child: const Text('确认')),
        ],
      ),
    );
    if (confirmed == true) {
      try {
        await _repo.cancelOrder(widget.orderNo);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('订单已取消')));
          _loadOrder();
        }
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('取消失败: $e')));
        }
      }
    }
  }

  Future<void> _confirmReceipt() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('确认收货'),
        content: const Text('确认已收到商品？确认后将自动完成分账。'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('取消')),
          TextButton(onPressed: () => Navigator.pop(context, true), child: const Text('确认收货')),
        ],
      ),
    );
    if (confirmed == true) {
      try {
        await _repo.confirmReceipt(widget.orderNo);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('已确认收货')));
          _loadOrder();
        }
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('操作失败: $e')));
        }
      }
    }
  }

  Future<void> _goToPay() async {
    final result = await Navigator.push<bool>(
      context,
      MaterialPageRoute(builder: (_) => PaymentScreen(order: _order!)),
    );
    if (result == true) {
      _loadOrder();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    if (_order == null) {
      return Scaffold(appBar: AppBar(title: const Text('订单详情')), body: const Center(child: Text('订单不存在')));
    }

    return Scaffold(
      appBar: AppBar(title: const Text('订单详情')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Status card
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: _statusBgColor(_order!.orderStatus),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                children: [
                  Icon(_statusIcon(_order!.orderStatus), size: 48, color: _statusTextColor(_order!.orderStatus)),
                  const SizedBox(height: 8),
                  Text(_statusLabel(_order!.orderStatus),
                      style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: _statusTextColor(_order!.orderStatus))),
                ],
              ),
            ),
            const SizedBox(height: 20),
            // Order info
            _sectionTitle('订单信息'),
            _infoCard('订单编号', _order!.orderNo),
            _infoCard('创建时间', _formatTime(_order!.createdAt)),
            if (_order!.paidAt != null) _infoCard('支付时间', _formatTime(_order!.paidAt!)),
            if (_order!.shippedAt != null) _infoCard('发货时间', _formatTime(_order!.shippedAt!)),
            if (_order!.paymentMethod != null) _infoCard('支付方式', _paymentLabel(_order!.paymentMethod!)),
            const SizedBox(height: 20),
            // Items
            _sectionTitle('商品列表'),
            ..._order!.items.map((item) => _itemCard(item)),
            const SizedBox(height: 20),
            // Total
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                const Text('合计: ', style: TextStyle(fontSize: 14)),
                Text('¥${_order!.totalAmount.toStringAsFixed(2)}',
                    style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Color(0xFFEF4444))),
              ],
            ),
          ],
        ),
      ),
      bottomNavigationBar: _buildBottomBar(),
    );
  }

  Widget _buildBottomBar() {
    if (_order!.orderStatus == 'pending') {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: SafeArea(
          child: Row(
            children: [
              OutlinedButton(onPressed: _cancelOrder, child: const Text('取消订单')),
              const SizedBox(width: 12),
              Expanded(
                child: FilledButton(
                  onPressed: _goToPay,
                  style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 14)),
                  child: const Text('去支付'),
                ),
              ),
            ],
          ),
        ),
      );
    }
    if (_order!.orderStatus == 'shipped') {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: SafeArea(
          child: FilledButton(
            onPressed: _confirmReceipt,
            style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 14)),
            child: const Text('确认收货'),
          ),
        ),
      );
    }
    return const SizedBox.shrink();
  }

  Widget _sectionTitle(String title) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
    );
  }

  Widget _infoCard(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          SizedBox(width: 70, child: Text(label, style: const TextStyle(fontSize: 13, color: Colors.grey))),
          Expanded(child: Text(value, style: const TextStyle(fontSize: 13))),
        ],
      ),
    );
  }

  Widget _itemCard(OrderItemModel item) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        title: Text('商品 ${item.productId.substring(0, 8)}...'),
        subtitle: Text('x${item.quantity}'),
        trailing: Text('¥${item.price.toStringAsFixed(2)}',
            style: const TextStyle(fontWeight: FontWeight.bold, color: Color(0xFFEF4444))),
      ),
    );
  }

  Color _statusBgColor(String status) {
    switch (status) {
      case 'pending': return Colors.orange.shade50;
      case 'paid': return Colors.blue.shade50;
      case 'shipped': return Colors.purple.shade50;
      case 'completed': return Colors.green.shade50;
      case 'cancelled': return Colors.grey.shade100;
      case 'refunded': return Colors.red.shade50;
      default: return Colors.grey.shade50;
    }
  }

  Color _statusTextColor(String status) {
    switch (status) {
      case 'pending': return Colors.orange.shade700;
      case 'paid': return Colors.blue.shade700;
      case 'shipped': return Colors.purple.shade700;
      case 'completed': return Colors.green.shade700;
      case 'cancelled': return Colors.grey;
      case 'refunded': return Colors.red;
      default: return Colors.grey;
    }
  }

  IconData _statusIcon(String status) {
    switch (status) {
      case 'pending': return Icons.pending_actions;
      case 'paid': return Icons.payment;
      case 'shipped': return Icons.local_shipping;
      case 'completed': return Icons.check_circle;
      case 'cancelled': return Icons.cancel;
      case 'refunded': return Icons.receipt_long;
      default: return Icons.help_outline;
    }
  }

  String _statusLabel(String status) {
    switch (status) {
      case 'pending': return '待付款';
      case 'paid': return '已付款';
      case 'shipped': return '已发货';
      case 'completed': return '已完成';
      case 'cancelled': return '已取消';
      case 'refunded': return '已退款';
      default: return status;
    }
  }

  String _paymentLabel(String method) {
    switch (method) {
      case 'wechat': return '微信支付';
      case 'alipay': return '支付宝';
      default: return method;
    }
  }

  String _formatTime(DateTime t) {
    return '${t.year}-${t.month.toString().padLeft(2, '0')}-${t.day.toString().padLeft(2, '0')} '
        '${t.hour.toString().padLeft(2, '0')}:${t.minute.toString().padLeft(2, '0')}';
  }
}
