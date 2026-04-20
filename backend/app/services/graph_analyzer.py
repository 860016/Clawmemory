import logging
from sqlalchemy.orm import Session
from app.models.knowledge import Entity, Relation
from app.services.license_service import is_feature_enabled

logger = logging.getLogger(__name__)


class GraphAnalyzer:
    """图分析引擎：路径查找、中心度、社区发现"""

    def __init__(self, db: Session):
        self.db = db

    def _build_graph(self):
        """构建 NetworkX 图"""
        try:
            import networkx as nx
        except ImportError:
            raise RuntimeError("networkx not installed")

        G = nx.DiGraph()
        entities = self.db.query(Entity).filter(Entity.user_id == 1).all()
        relations = self.db.query(Relation).filter(Relation.user_id == 1).all()

        for e in entities:
            G.add_node(e.id, name=e.name, type=e.entity_type, description=e.description or "")
        for r in relations:
            G.add_edge(r.source_id, r.target_id,
                       type=r.relation_type, weight=r.weight or 1.0)
        return G

    def shortest_path(self, source_id: int, target_id: int) -> list[dict]:
        """两实体间最短路径"""
        import networkx as nx
        G = self._build_graph()
        try:
            path = nx.shortest_path(G, source_id, target_id, weight="weight")
            result = []
            for i, node_id in enumerate(path):
                node_data = {
                    "id": node_id,
                    "name": G.nodes[node_id]["name"],
                    "type": G.nodes[node_id]["type"],
                }
                if i > 0:
                    edge_data = G.edges[path[i-1], node_id]
                    node_data["relation"] = edge_data.get("type", "")
                result.append(node_data)
            return result
        except (nx.NetworkXNoPath, nx.NodeNotFound):
            return []

    def centrality_analysis(self) -> list[dict]:
        """中心度分析 — 找出核心实体"""
        import networkx as nx
        G = self._build_graph()
        if G.number_of_nodes() == 0:
            return []

        degree = nx.degree_centrality(G)
        betweenness = nx.betweenness_centrality(G)
        pagerank = nx.pagerank(G)

        results = []
        for node_id in G.nodes:
            results.append({
                "id": node_id,
                "name": G.nodes[node_id]["name"],
                "type": G.nodes[node_id]["type"],
                "degree_centrality": round(degree.get(node_id, 0), 4),
                "betweenness_centrality": round(betweenness.get(node_id, 0), 4),
                "pagerank": round(pagerank.get(node_id, 0), 4),
            })
        return sorted(results, key=lambda x: x["pagerank"], reverse=True)

    def community_detection(self) -> list[dict]:
        """社区发现 — 聚类相关实体"""
        import networkx as nx
        G = self._build_graph().to_undirected()
        if G.number_of_nodes() == 0:
            return []

        try:
            communities = nx.community.louvain_communities(G, seed=42)
        except Exception:
            # Fallback: connected components
            communities = list(nx.connected_components(G))

        result = []
        for i, community in enumerate(communities):
            entities = []
            for node_id in community:
                if node_id in G.nodes:
                    entities.append({
                        "id": node_id,
                        "name": G.nodes[node_id]["name"],
                        "type": G.nodes[node_id]["type"],
                    })
            result.append({
                "community_id": i,
                "entities": entities,
                "size": len(entities),
            })
        return sorted(result, key=lambda x: x["size"], reverse=True)

    def entity_neighbors(self, entity_id: int, depth: int = 2) -> dict:
        """实体邻居探索（BFS）"""
        import networkx as nx
        G = self._build_graph()
        if entity_id not in G:
            return {"error": "Entity not found"}

        subgraph = nx.ego_graph(G, entity_id, radius=depth)

        nodes = []
        for node_id in subgraph.nodes:
            nodes.append({
                "id": node_id,
                "name": subgraph.nodes[node_id]["name"],
                "type": subgraph.nodes[node_id]["type"],
                "is_center": node_id == entity_id,
            })

        edges = []
        for src, tgt, data in subgraph.edges(data=True):
            edges.append({
                "source_id": src,
                "target_id": tgt,
                "relation_type": data.get("type", ""),
            })

        return {
            "center_entity": {
                "id": entity_id,
                "name": G.nodes[entity_id]["name"],
                "type": G.nodes[entity_id]["type"],
            },
            "depth": depth,
            "nodes": nodes,
            "edges": edges,
            "total_nodes": len(nodes),
            "total_edges": len(edges),
        }

    def get_graph_stats(self) -> dict:
        """图谱统计信息"""
        import networkx as nx
        G = self._build_graph()

        if G.number_of_nodes() == 0:
            return {
                "total_entities": 0,
                "total_relations": 0,
                "density": 0,
                "avg_degree": 0,
                "connected_components": 0,
            }

        undirected = G.to_undirected()
        return {
            "total_entities": G.number_of_nodes(),
            "total_relations": G.number_of_edges(),
            "density": round(nx.density(G), 4),
            "avg_degree": round(sum(dict(G.degree()).values()) / G.number_of_nodes(), 2),
            "connected_components": nx.number_connected_components(undirected),
        }
