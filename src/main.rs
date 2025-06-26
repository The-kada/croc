mod ecs;

fn main() {
    let my_entity = ecs::entity::Entity { entity_id: 4 };
    println!("{:?}", my_entity);
}
