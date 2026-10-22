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

export interface TreeLayout {
  width: number;
  height: number;
  nodes: TreeNodeLayout[];
  sink?: TreeBox;
  edges: string[];
  sinkEdges: string[];
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
  const edges: string[] = [];
  const sinkEdges: string[] = [];

  const visit = (placement: Placement) => {
    const box = toBox(placement.column, placement.depth);
    nodes.push({ item: placement.item, ...box });
    for (const child of placement.children) {
      edges.push(elbow(box, toBox(child.column, child.depth)));
      visit(child);
    }
    if (sink && placement.sinkColumn !== undefined) {
      const stub = toBox(placement.sinkColumn, sinkDepth);
      sinkEdges.push(
        `M${centerX(box)} ${bottomY(box)}V${elbowY(box)}H${centerX(stub)}V${sink.y - ROW_GAP / 2}H${centerX(sink)}V${sink.y}`,
      );
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
