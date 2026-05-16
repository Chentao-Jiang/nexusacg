import 'package:flutter/material.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/screens/orders/order_detail_screen.dart';

class OrdersScreen extends StatefulWidget {
  const OrdersScreen({super.key});

  @override
  State<OrdersScreen> createState() => _OrdersScreenState();
}

class _OrdersScreenState extends State<OrdersScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  final _repo = OrderRepository();
  List<OrderModel> _orders = [];
  bool _loading = true;
  int _page = 1;
  bool _hasMore = true;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 5, vsync: this);
    _tabController.addListener(() {
      if (!_tabController.indexIsChanging) {
        _loadOrders();
      }
    });
    _loadOrders();
  }

  String _getFilterStatus() {
    switch (_tabController.index) {
      case 0: return ''; // All
      case 1: return 'pending';
      case 2: return 'paid';
      case 3: return 'shipped';
      case 4: return 'completed';
      default: return '';
    }
  }

  Future<void> _loadOrders() async {
    setState(() => _loading = true);
    _page = 1;
    try {
      final result = await _repo.getOrders(page: 1, status: _getFilterStatus());
      setState(() {
        _orders = result.items;
        _hasMore = result.items.length >= 20;
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  Future<void> _loadMore() async {
    if (!_hasMore) return;
    final nextPage = _page + 1;
    try {
      final result = await _repo.getOrders(page: nextPage, status: _getFilterStatus());
      setState(() {
        _orders.addAll(result.items);
        _page = nextPage;
        _hasMore = result.items.length >= 20;
      });
    } catch (e) {
      // Ignore
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('我的订单'),
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          tabs: const [
            Tab(text: '全部'),
            Tab(text: '待付款'),
            Tab(text: '已付款'),
            Tab(text: '已发货'),
            Tab(text: '已完成'),
          ],
        ),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _orders.isEmpty
              ? const Center(child: Text('暂无订单'))
              : RefreshIndicator(
                  onRefresh: _loadOrders,
                  child: NotificationListener<ScrollNotification>(
                    onNotification: (notification) {
                      if (notification is ScrollEndNotification &&
                          notification.metrics.pixels >= notification.metrics.maxScrollExtent * 0.8) {
                        _loadMore();
                      }
                      return false;
                    },
                    child: ListView.builder(
                      padding: const EdgeInsets.all(12),
                      itemCount: _orders.length,
                      itemBuilder: (context, index) => _OrderCard(_orders[index]),
                    ),
                  ),
                ),
    );
  }
}

class _OrderCard extends StatelessWidget {
  final OrderModel order;
  const _OrderCard(this.order);

  @override
  Widget build(BuildContext context) {
    final statusColor = _statusColor(order.orderStatus);
    final statusText = _statusLabel(order.orderStatus);

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: () => Navigator.push(
          context,
          MaterialPageRoute(builder: (_) => OrderDetailScreen(orderNo: order.orderNo)),
        ),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Text('订单号', style: TextStyle(fontSize: 12, color: Colors.grey)),
                  const Spacer(),
                  Text(statusText,
                      style: TextStyle(fontSize: 13, color: statusColor, fontWeight: FontWeight.w500)),
                ],
              ),
              const SizedBox(height: 8),
              Text(order.orderNo, style: const TextStyle(fontSize: 12, color: Colors.grey)),
              if (order.items.isNotEmpty) ...[
                const SizedBox(height: 8),
                Text('${order.items.first.productName} ${order.items.length > 1 ? '等${order.items.length}件商品' : ''}',
                    style: const TextStyle(fontSize: 14)),
              ],
              const SizedBox(height: 8),
              Row(
                children: [
                  const Spacer(),
                  const Text('合计: ', style: TextStyle(fontSize: 13, color: Colors.grey)),
                  Text('¥${order.totalAmount.toStringAsFixed(2)}',
                      style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Color(0xFFEF4444))),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Color _statusColor(String status) {
    switch (status) {
      case 'pending': return Colors.orange;
      case 'paid': return Colors.blue;
      case 'shipped': return Colors.purple;
      case 'completed': return Colors.green;
      case 'cancelled': return Colors.grey;
      case 'refunded': return Colors.red;
      default: return Colors.grey;
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
}

// Extension for displaying product info in order cards
extension on OrderItemModel {
  String get productName => '商品 $productId';
}
