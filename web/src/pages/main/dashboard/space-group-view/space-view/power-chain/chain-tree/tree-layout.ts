import type { ChainTreeNode } from "@/lib/power-chain";

export const NODE_WIDTH = 184;
export const NODE_HEIGHT = 48;

const COLUMN_GAP = 24;
const ROW_GAP = 40;
const COLUMN_PITCH = NODE_WIDTH + COLUMN_GAP;
const ROW_PITCH = NODE_HEIGHT + ROW_GAP;

export interface TreeBox {
  x: number;
  y: number;
}

export interface TreeNodeLayout extends TreeBox {
  item: ChainTreeNode;
}

export interface TreeEdge {
  path: string;
  hitPath: string;
  parent: ChainTreeNode;
  child?: ChainTreeNode;
}

export interface TreeLayout {
  width: number;
  height: number;
  nodes: TreeNodeLayout[];
  sink?: TreeBox;
  edges: TreeEdge[];
  sinkEdges: TreeEdge[];
}

interface Placement {
  item: ChainTreeNode;
  column: number;
  depth: number;
  children: Placement[];
  sinkColumn?: number;
}

const toBox = (column: number, depth: number): TreeBox => ({
  x: column * COLUMN_PITCH,
  y: depth * ROW_PITCH,
});

const centerX = ({ x }: TreeBox) => x + NODE_WIDTH / 2;

const bottomY = ({ y }: TreeBox) => y + NODE_HEIGHT;

const elbowY = (box: TreeBox) => bottomY(box) + ROW_GAP / 2;

const elbow = (from: TreeBox, to: TreeBox) =>
  `M${centerX(from)} ${bottomY(from)}V${elbowY(from)}H${centerX(to)}V${to.y}`;

export const layoutTree = (
  roots: ChainTreeNode[],
  hasSink: boolean,
): TreeLayout => {
  let columns = 0;
  let deepest = 0;
  const sinkColumns: number[] = [];

  const place = (item: ChainTreeNode, depth: number): Placement => {
    deepest = Math.max(deepest, depth);
    const children = item.children.map((child) => place(child, depth + 1));
    const spans = children.map(({ column }) => column);
    const sinkColumn = hasSink && item.feedsSink ? columns++ : undefined;
    if (sinkColumn !== undefined) {
      spans.push(sinkColumn);
      sinkColumns.push(sinkColumn);
    }
    const column =
      spans.length === 0 ? columns++ : (spans[0] + spans[spans.length - 1]) / 2;
    return { item, column, depth, children, sinkColumn };
  };

  const placements = roots.map((root) => place(root, 0));
  const sinkDepth = deepest + 1;
  const sink =
    sinkColumns.length === 0
      ? undefined
      : toBox(
          sinkColumns.reduce((sum, column) => sum + column, 0) /
            sinkColumns.length,
          sinkDepth,
        );

  const nodes: TreeNodeLayout[] = [];
  const edges: TreeEdge[] = [];
  const sinkEdges: TreeEdge[] = [];

  const visit = (placement: Placement) => {
    const box = toBox(placement.column, placement.depth);
    nodes.push({ item: placement.item, ...box });
    for (const child of placement.children) {
      const childBox = toBox(child.column, child.depth);
      edges.push({
        path: elbow(box, childBox),
        hitPath: `M${centerX(childBox)} ${elbowY(box)}V${childBox.y}`,
        parent: placement.item,
        child: child.item,
      });
      visit(child);
    }
    if (sink && placement.sinkColumn !== undefined) {
      const stub = toBox(placement.sinkColumn, sinkDepth);
      const mergeY = sink.y - ROW_GAP / 2;
      sinkEdges.push({
        path: `M${centerX(box)} ${bottomY(box)}V${elbowY(box)}H${centerX(stub)}V${mergeY}H${centerX(sink)}V${sink.y}`,
        hitPath: `M${centerX(stub)} ${elbowY(box)}V${mergeY}`,
        parent: placement.item,
      });
    }
  };
  placements.forEach(visit);

  return {
    width: Math.max(columns - 1, 0) * COLUMN_PITCH + NODE_WIDTH,
    height: (sink ? sinkDepth : deepest) * ROW_PITCH + NODE_HEIGHT,
    nodes,
    sink,
    edges,
    sinkEdges,
  };
};
