import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/screens/orders/checkout_screen.dart';

class ProductDetailScreen extends StatefulWidget {
  final ProductModel product;
  const ProductDetailScreen({super.key, required this.product});

  @override
  State<ProductDetailScreen> createState() => _ProductDetailScreenState();
}

class _ProductDetailScreenState extends State<ProductDetailScreen> {
  late ProductModel _product;
  int _quantity = 1;
  final _repo = ProductRepository();
  int _currentImageIndex = 0;

  @override
  void initState() {
    super.initState();
    _product = widget.product;
    _loadDetail();
  }

  Future<void> _loadDetail() async {
    setState(() {});
    try {
      final detail = await _repo.getProduct(_product.id);
      if (detail != null && mounted) {
        setState(() => _product = detail);
      }
    } catch (e) {
      // Ignore, use cached data
    } finally {
      if (mounted) setState(() {});
    }
  }

  void _goToCheckout() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => CheckoutScreen(productId: _product.id, quantity: _quantity, product: _product),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 350,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              background: _product.images.isNotEmpty
                  ? PageView.builder(
                      itemCount: _product.images.length,
                      onPageChanged: (i) => setState(() => _currentImageIndex = i),
                      itemBuilder: (context, index) {
                        return CachedNetworkImage(
                          imageUrl: _product.images[index],
                          fit: BoxFit.cover,
                          width: double.infinity,
                          errorWidget: (_, __, ___) => Container(
                            color: Colors.grey.shade200,
                            child: const Center(child: Icon(Icons.image, size: 64, color: Colors.grey)),
                          ),
                        );
                      },
                    )
                  : Container(color: Colors.grey.shade200, child: const Center(child: Icon(Icons.image, size: 64, color: Colors.grey))),
            ),
          ),
          if (_product.images.length > 1)
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: List.generate(_product.images.length, (i) {
                    return Container(
                      margin: const EdgeInsets.symmetric(horizontal: 3),
                      width: 6,
                      height: 6,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: i == _currentImageIndex ? colorScheme.primary : Colors.grey.shade400,
                      ),
                    );
                  }),
                ),
              ),
            ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Price
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        '¥${_product.price.toStringAsFixed(2)}',
                        style: TextStyle(fontSize: 28, fontWeight: FontWeight.bold, color: colorScheme.error),
                      ),
                      const SizedBox(width: 12),
                      if (_product.originalPrice != null)
                        Text(
                          '¥${_product.originalPrice!.toStringAsFixed(2)}',
                          style: const TextStyle(fontSize: 14, color: Colors.grey, decoration: TextDecoration.lineThrough),
                        ),
                      const Spacer(),
                      if (_product.stock <= 0)
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                          decoration: BoxDecoration(
                            color: Colors.grey.shade200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: const Text('已售罄', style: TextStyle(fontSize: 12, color: Colors.grey)),
                        ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  // Name
                  Text(_product.name, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                  if (_product.animeName != null) ...[
                    const SizedBox(height: 6),
                    Text('作品: ${_product.animeName}', style: TextStyle(fontSize: 13, color: colorScheme.secondary)),
                  ],
                  if (_product.characterName != null) ...[
                    const SizedBox(height: 4),
                    Text('角色: ${_product.characterName}', style: TextStyle(fontSize: 13, color: colorScheme.secondary)),
                  ],
                  const SizedBox(height: 16),
                  const Divider(),
                  const SizedBox(height: 8),
                  // Info grid
                  _infoRow('分区', _product.zone == 'cosplay' ? 'Cosplay专区' : '周边专区'),
                  _infoRow('来源', _sourceTypeLabel(_product.sourceType)),
                  _infoRow('库存', '${_product.stock}件'),
                  if (_product.tags.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 6,
                      children: _product.tags.map((t) => Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: colorScheme.primary.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(t, style: TextStyle(fontSize: 12, color: colorScheme.primary)),
                      )).toList(),
                    ),
                  ],
                  const SizedBox(height: 20),
                  const Divider(),
                  const SizedBox(height: 8),
                  // Description
                  const Text('商品详情', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 12),
                  Text(_product.description.isEmpty ? '暂无详情' : _product.description,
                      style: const TextStyle(fontSize: 14, height: 1.8)),
                  const SizedBox(height: 24),
                ],
              ),
            ),
          ),
        ],
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
              if (_product.stock > 0)
                _quantitySelector(context),
              const SizedBox(width: 12),
              Expanded(
                child: FilledButton(
                  onPressed: _product.stock > 0 ? _goToCheckout : null,
                  style: FilledButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  child: const Text('立即购买'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          SizedBox(width: 60, child: Text(label, style: const TextStyle(fontSize: 13, color: Colors.grey))),
          Text(value, style: const TextStyle(fontSize: 13)),
        ],
      ),
    );
  }

  String _sourceTypeLabel(String type) {
    switch (type) {
      case 'self_made': return '原创';
      case 'official': return '官方';
      case 'agent': return '代理';
      default: return type;
    }
  }

  Widget _quantitySelector(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.grey.shade300),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          InkWell(
            onTap: _quantity > 1 ? () => setState(() => _quantity--) : null,
            borderRadius: const BorderRadius.horizontal(left: Radius.circular(7)),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              child: Icon(Icons.remove, size: 18,
                  color: _quantity > 1 ? colorScheme.primary : Colors.grey),
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Text('$_quantity', style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
          ),
          InkWell(
            onTap: _quantity < _product.stock ? () => setState(() => _quantity++) : null,
            borderRadius: const BorderRadius.horizontal(right: Radius.circular(7)),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              child: Icon(Icons.add, size: 18,
                  color: _quantity < _product.stock ? colorScheme.primary : Colors.grey),
            ),
          ),
        ],
      ),
    );
  }
}
