import 'package:flutter/material.dart';
import 'package:nexusacg/presentation/screens/settings/legal_content.dart';

class LegalPage extends StatelessWidget {
  final String title;
  final String content;
  const LegalPage({super.key, required this.title, required this.content});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Text(content, style: const TextStyle(fontSize: 14, height: 1.8)),
      ),
    );
  }
}
