import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/presentation/blocs/product/product_bloc.dart';
import 'package:nexusacg/presentation/blocs/product/product_state.dart';
import 'package:nexusacg/presentation/screens/products/product_detail_screen.dart';

class ProductScreen extends StatefulWidget {
  final int initialTabIndex;
  const ProductScreen({super.key, this.initialTabIndex = 0});

  @override
  State<ProductScreen> createState() => _ProductScreenState();
}

class _ProductScreenState extends State<ProductScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this, initialIndex: widget.initialTabIndex);
    _tabController.addListener(() {
      if (!_tabController.indexIsChanging) {
        final zone = _tabController.index == 0 ? 'cosplay' : 'peripheral';
        context.read<ProductBloc>().add(ProductLoadRequested(zone: zone));
      }
    });
    // Load initial tab
    final zone = widget.initialTabIndex == 0 ? 'cosplay' : 'peripheral';
    context.read<ProductBloc>().add(ProductLoadRequested(zone: zone));
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('商城'),
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Cosplay专区'),
            Tab(text: '周边专区'),
          ],
        ),
      ),
      body: BlocBuilder<ProductBloc, ProductState>(
        builder: (context, state) {
          if (state is ProductLoading) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is ProductLoaded) {
            return state.products.isEmpty
                ? const Center(child: Text('暂无商品'))
                : GridView.builder(
                    padding: const EdgeInsets.all(12),
                    gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 2,
                      childAspectRatio: 0.7,
                      crossAxisSpacing: 12,
                      mainAxisSpacing: 12,
                    ),
                    itemCount: state.products.length,
                    itemBuilder: (context, index) => _ProductCard(state.products[index]),
                  );
          }
          return const Center(child: Text('加载失败'));
        },
      ),
    );
  }
}

class _ProductCard extends StatelessWidget {
  final ProductModel product;
  const _ProductCard(this.product);

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => ProductDetailScreen(product: product)),
          );
        },
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Container(
                color: Colors.grey.shade200,
                child: product.images.isNotEmpty
                    ? Image.network(product.images.first, fit: BoxFit.cover, errorBuilder: (_, __, ___) => const Center(child: Icon(Icons.image, size: 48, color: Colors.grey)))
                    : const Center(child: Icon(Icons.image, size: 48, color: Colors.grey)),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    product.name,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500),
                  ),
                  const SizedBox(height: 4),
                  if (product.animeName != null)
                    Text(product.animeName!, style: const TextStyle(fontSize: 11, color: Colors.grey)),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      Text('¥${product.price.toStringAsFixed(2)}',
                          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Color(0xFFEF4444))),
                      if (product.originalPrice != null)
                        Text('¥${product.originalPrice!.toStringAsFixed(2)}',
                            style: const TextStyle(fontSize: 11, color: Colors.grey, decoration: TextDecoration.lineThrough)),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
